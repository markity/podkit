package json_struct

import (
	"encoding/json"
	"os"
)

type ContainerInfo struct {
	ContainerID        int    `json:"container_id"`
	ContainerImageName string `json:"container_image_name"`
}

type RunningInfoStruct struct {
	ContainerIDCount int              `json:"container_id_count"`
	ContainerRunning []*ContainerInfo `json:"container_running"`
	ContainerStopped []*ContainerInfo `json:"container_stopped"`
}

func (ris *RunningInfoStruct) ParseFromFile(filePath string) error {
	bs, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, ris)
	if err != nil {
		return err
	}

	return nil
}

func (ris *RunningInfoStruct) MustMarshalToBytes() []byte {
	b, err := json.Marshal(ris)
	if err != nil {
		panic(err)
	}
	return b
}

type ImageInfoStruct struct {
	ImageTarFilename map[string]string `json:"image_tar_filename"`
}

func (ims *ImageInfoStruct) ParseFromFile(filePath string) error {
	bs, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, ims)
	if err != nil {
		return err
	}

	return nil
}
