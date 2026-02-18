package mta

import "time"

type FileMetadata struct {
	Id              string `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Size            int32  `json:"size,omitempty"`
	Digest          string `json:"digest,omitempty"`
	DigestAlgorithm string `json:"digestAlgorithm,omitempty"`
	Space           string `json:"space,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
}

type UploadStatus struct {
	Status string       `json:"status,omitempty"`
	File   FileMetadata `json:"file"`
	MtaId  string       `json:"mta_id,omitempty"`
	Error  string       `json:"error,omitempty"`
}

type ProcessType struct {
	Name string `json:"name,omitempty"`
}

type Operation struct {
	ProcessId    string         `json:"processId,omitempty"`
	ProcessType  string         `json:"processType,omitempty"`
	StartedAt    string         `json:"startedAt,omitempty"`
	EndedAt      string         `json:"endedAt,omitempty"`
	SpaceId      string         `json:"spaceId,omitempty"`
	MtaId        string         `json:"mtaId,omitempty"`
	Namespace    string         `json:"namespace,omitempty"`
	User         string         `json:"user,omitempty"`
	AcquiredLock bool           `json:"acquiredLock,omitempty"`
	State        string         `json:"state,omitempty"`
	ErrorType    string         `json:"errorType,omitempty"`
	Messages     []Message      `json:"messages,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
}

type Mta struct {
	Metadata *Metadata `json:"metadata,omitempty"`
	Modules  []Module  `json:"modules,omitempty"`
	Services []string  `json:"services,omitempty"`
}

type Module struct {
	ModuleName            string    `json:"moduleName,omitempty"`
	AppName               string    `json:"appName,omitempty"`
	CreatedOn             time.Time `json:"createdOn"`
	UpdatedOn             time.Time `json:"updatedOn"`
	ProvidedDendencyNames []string  `json:"providedDendencyNames,omitempty"`
	Services              []string  `json:"services,omitempty"`
	Uris                  []string  `json:"uris,omitempty"`
}

type Metadata struct {
	Id        string `json:"id,omitempty"`
	Version   string `json:"version,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type Message struct {
	Id        int64  `json:"id,omitempty"`
	Text      string `json:"text,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Type_     string `json:"type,omitempty"`
}

type Log struct {
	Id           string    `json:"id,omitempty"`
	LastModified time.Time `json:"lastModified"`
	Content      string    `json:"content,omitempty"`
	Size         int64     `json:"size,omitempty"`
	DisplayName  string    `json:"displayName,omitempty"`
	Description  string    `json:"description,omitempty"`
	ExternalInfo string    `json:"externalInfo,omitempty"`
}

type FileUrl struct {
	FileUrl string `json:"file_url,omitempty"`
}
