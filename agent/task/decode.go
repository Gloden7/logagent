package task

import (
	"bufio"
	"fmt"
	"logagent/util"
	"strconv"
)

func parseFacilitySeverity(p []byte) (int, int, error) {
	fs, err := strconv.Atoi(util.Bytes2str(p))
	if err != nil {
		return 0, 0, err
	}
	if fs > 191 || fs < 0 {
		return 0, 0, fmt.Errorf("Unknown facility name")
	}
	return fs >> 3, fs & 0x07, nil
}

func decode(reader *bufio.Reader, end byte) (map[string]interface{}, error) {
	priorityBytes, err := reader.ReadBytes('>')
	if err != nil {
		return nil, err
	}
	facility, severity, err := parseFacilitySeverity(priorityBytes[1 : len(priorityBytes)-1])
	if err != nil {
		return nil, err
	}
	msgBytes, err := reader.ReadBytes(end)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"facility": facility,
		"severity": severity,
		"message":  util.Bytes2str(msgBytes[:len(msgBytes)-1]),
	}, nil
}
