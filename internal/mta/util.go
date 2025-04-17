package mta

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultDescriptorPath string = "META-INF/mtad.yaml"
	FinishedState         string = "FINISHED"
	AbortedState          string = "ABORTED"
)

type MtaDescriptor struct {
	SchemaVersion string `yaml:"_schema-version,omitempty"`
	ID            string `yaml:"ID,omitempty"`
	Version       string `yaml:"version,omitempty"`
	Namespace     string `yaml:"namespace,omitempty"`
}

// ref - https://github.com/cloudfoundry/multiapps-cli-plugin/blob/v3.2.2/util/archive_handler.go
// GetMtaDescriptorFromArchive retrieves MTA ID from MTA archive.
func GetMtaDescriptorFromArchive(mtaArchiveFilePath string) (MtaDescriptor, error) {
	mtaArchiveReader, err := zip.OpenReader(mtaArchiveFilePath)
	if err != nil {
		return MtaDescriptor{}, err
	}
	defer mtaArchiveReader.Close()

	descriptorFile := findMtaDescriptorFile(mtaArchiveReader.File)
	if descriptorFile == nil {
		return MtaDescriptor{}, errors.New("could not get a valid mta descriptor from archive")
	}

	descriptorBytes, err := readZipFile(descriptorFile)
	if err != nil {
		return MtaDescriptor{}, err
	}

	var descriptor MtaDescriptor
	if err = yaml.Unmarshal(descriptorBytes, &descriptor); err != nil {
		return MtaDescriptor{}, err
	}

	if descriptor.ID != "" {
		return descriptor, nil
	}
	return MtaDescriptor{}, errors.New("could not get a valid mta descriptor from archive")
}

func findMtaDescriptorFile(files []*zip.File) *zip.File {
	for _, file := range files {
		if file.Name == defaultDescriptorPath {
			return file
		}
	}
	return nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// ref - https://github.com/cloudfoundry/multiapps-cli-plugin/blob/v3.2.2/commands/deploy_command.go
// CheckOngoingOperation checks for ongoing operation for mta with the specified id and tries to abort it.
func CheckOngoingOperation(ctx context.Context, client *APIClient, mtaId string, namespace string, spaceGuid string) (bool, error) {
	// Check if there is an ongoing operation for this MTA ID
	ongoingOperation, err := findOngoingOperation(ctx, mtaId, namespace, client, spaceGuid)
	if err != nil {
		return false, err
	}
	if ongoingOperation != nil {
		// Abort the conflicting process
		operationId, _, err := client.DefaultApi.ExecuteOperationAction(ctx, spaceGuid, ongoingOperation.ProcessId, "abort")
		if err != nil {
			return false, err
		} else {
			_, err = PollMtaOperation(ctx, client, spaceGuid, operationId, AbortedState)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// FindOngoingOperation finds ongoing operation for mta with the specified id.
func findOngoingOperation(ctx context.Context, mtaID string, namespace string, client *APIClient, spaceGuid string) (*Operation, error) {
	activeStatesList := []string{"RUNNING", "ERROR", "ACTION_REQUIRED"}
	getOptions := &DefaultApiGetMtaOperationsOpts{
		MtaId: &mtaID,
		State: activeStatesList,
	}
	ongoingOperations, _, err := client.DefaultApi.GetMtaOperations(ctx, spaceGuid, getOptions)
	if err != nil {
		return nil, fmt.Errorf("could not get ongoing operations for multi-target app %s: %s", mtaID, err)
	}
	for _, ongoingOperation := range ongoingOperations {
		isConflicting := isConflicting(ongoingOperation, mtaID, namespace, spaceGuid)
		if isConflicting {
			return &ongoingOperation, nil
		}
	}
	return nil, nil
}

func isConflicting(operation Operation, mtaID string, namespace string, spaceGuid string) bool {
	return operation.MtaId == mtaID &&
		operation.SpaceId == spaceGuid &&
		operation.Namespace == namespace &&
		operation.AcquiredLock
}

// Keeps polling the MTA operation by its ID for completion.
func PollMtaOperation(ctx context.Context, client *APIClient, spaceGuid string, operationId string, targetState string) (string, error) {

	var (
		operationResponse Operation
		err               error
	)
	for operationState := "RUNNING"; operationState != targetState; {
		time.Sleep(2 * time.Second)
		operationResponse, _, err = client.DefaultApi.GetMtaOperation(ctx, spaceGuid, operationId, "messages")
		if err != nil {
			return "", err
		}
		operationState = operationResponse.State
		if operationState == "ERROR" {
			if messageCount := len(operationResponse.Messages); messageCount > 0 {
				return messagesToString(operationResponse.Messages), fmt.Errorf("last message %s", operationResponse.Messages[messageCount-1].Text)
			}
			return "", fmt.Errorf("Operation failed with errorType %s", operationResponse.ErrorType)
		}
	}
	return messagesToString(operationResponse.Messages), nil
}

// ref - https://github.com/cloudfoundry/go-cfclient/blob/main/internal/http/response.go
// returns the operationID if specified in the Location response header.
func decodeOperationJobID(resp *http.Response) (string, error) {
	// Return empty if the response is nil
	if resp == nil {
		return "", fmt.Errorf("no response obtained")
	}

	// If we succeed, return the operation GUID, otherwise pass on to decode error.
	location, err := resp.Location()
	if err != nil {
		// Return empty if there's an error (e.g., no Location header)
		return "", err
	}

	operationId := ""
	// Split the path in the URL and check for the 'jobs' segment
	parts := strings.Split(location.Path, "/")
	numParts := len(parts)
	// Ensure 'operations' is the second last element and return the last element as job ID
	if numParts >= 2 && parts[numParts-2] == "operations" || parts[numParts-2] == "jobs" {
		operationId = strings.Split(parts[numParts-1], "?")[0]
	} else {
		err = fmt.Errorf("did not find operation or job id in location header")
	}
	return operationId, err
}

// Keeps polling the MTA job by its ID for completion.
func PollMtaJob(ctx context.Context, client *APIClient, spaceGuid string, jobId string, targetState string, xInstance string, namespace string) (jobResponse UploadStatus, err error) {
	for jobState := "RUNNING"; jobState != targetState; {
		time.Sleep(2 * time.Second)
		jobResponse, _, err = client.DefaultApi.GetAsyncUploadJob(ctx, spaceGuid, jobId, xInstance, namespace)
		if err != nil {
			return jobResponse, err
		}
		jobState = jobResponse.Status
		if jobState == "ERROR" {
			return jobResponse, fmt.Errorf("upload job failed with %s", jobResponse.Error)
		}
	}
	return jobResponse, nil
}

// Combines multiple messages into a single string.
func messagesToString(messages []Message) (combinedMessage string) {
	for _, message := range messages {
		combinedMessage = combinedMessage + "\n" + message.Text
	}
	return combinedMessage
}
