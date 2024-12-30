package provider

import (
	"fmt"
	"strconv"
	"strings"
)

func portRangeParse(portRange string) (start int, end int, err error) {
	portRangeSplit := strings.Split(portRange, "-")
	if len(portRangeSplit) > 2 {
		return 0, 0, fmt.Errorf("invalid range")
	}
	start, err = strconv.Atoi(portRangeSplit[0])
	if err != nil {
		return 0, 0, err
	}
	if len(portRangeSplit) == 1 {
		return start, start, nil
	}
	end, err = strconv.Atoi(portRangeSplit[1])
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}
