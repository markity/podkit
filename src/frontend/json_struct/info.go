package json_struct

import (
	"encoding/json"
	"os"
)

type ContainerInfo struct {
	ContainerID        int    `json:"container_id"`
	ContainerImageName string `json:"container_image_name"`
	IP                 string `json:"ip"`
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

func (ris *RunningInfoStruct) WriteToFile(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(ris.MustMarshalToBytes())
	return err
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
