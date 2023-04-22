package stablediffusion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/M-Ro/aurora-ai/api"
	"github.com/M-Ro/aurora-ai/config"
	"github.com/M-Ro/aurora-ai/internal/gradio"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type OnCompleteFunc func(images []bytes.Reader, err error)

func Run(parameters *ParameterSet, onComplete OnCompleteFunc) error {
	s := gradio.GetSession()
	apiConn := gradio.NewAPIConnection()
	host := viper.GetString("stable_diffusion.host")

	err := apiConn.Connect(host)
	if err != nil {
		return err
	}

	err = runSocketHandler(apiConn, parameters, s, onComplete)

	return err
}

func runSocketHandler(
	conn *gradio.APIConnection,
	parameters *ParameterSet,
	s *gradio.Session,
	onComplete OnCompleteFunc,
) error {
	sendQueue := make(chan string)
	done := make(chan struct{})
	go func() {
		defer close(done)

		for {
			_, message, err := conn.Ws.ReadMessage()
			if err != nil {
				logrus.Error("ws read: ", err)
				return
			}

			respPacket := SdResponsePacket{}
			err = json.Unmarshal(message, &respPacket)
			if err != nil {
				logrus.Error("Failed to unmarshal response: ", err)
				return
			}

			switch respPacket.Message {
			case api.MsgSendHash:
				onServerRequestHash(s, sendQueue)
			case api.MsgSendData:
				onServerRequestData(s, parameters, sendQueue)
			case api.MsgProcessCompleted:
				images, err := fetchImagesFromSd(respPacket)
				onComplete(images, err)
				return
			case api.MsgProcessGenerating:
				continue
			case api.MsgEstimation:
				continue
			}
		}
	}()

	for {
		select {
		case <-done:
			return nil
		case m := <-sendQueue:
			logrus.Info("Send Message " + m)
			err := conn.Ws.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				logrus.Error("write:", err)
				return nil
			}
		}
	}
}

func onServerRequestHash(session *gradio.Session, sendQueue chan string) {
	bytes, err := json.Marshal(api.SendHashRequest{
		SessionHash: session.SessionHash,
		FnIndex:     config.FnDoTheThing,
	})

	if err != nil {
		logrus.Error("onServerRequestHash: Failed marshalling")
		return
	}

	sendQueue <- string(bytes)
}
func onServerRequestData(session *gradio.Session, parameters *ParameterSet, sendQueue chan string) {
	bytes, err := json.Marshal(api.SendInferenceDataRequest{
		SessionHash: session.SessionHash,
		FnIndex:     config.FnDoTheThing,
		Data:        parameters,
	})

	if err != nil {
		logrus.Error("onServerRequestData: Failed to marshal")
		return
	}

	sendQueue <- string(bytes)
}

var (
	ErrFailureOnGeneration = errors.New("Received upstream error from SD interface")
	ErrNoImage             = errors.New("Upstream did not provide any generated images")
	ErrFetchImages         = errors.New("Failed to fetch one or more images from upstream")
	ErrNoOutput            = errors.New("No output block found in received packet")
	ErrFailedParsing       = errors.New("Failed parsing output data block")
)

func fetchImagesFromSd(response SdResponsePacket) ([]bytes.Reader, error) {
	if !*response.Success {
		return []bytes.Reader{}, ErrFailureOnGeneration
	}

	if response.Output == nil {
		return []bytes.Reader{}, ErrNoOutput
	}

	if len(response.Output.Data.Images) == 0 {
		return []bytes.Reader{}, ErrNoImage
	}

	// Fetch the images into buffers and attach readers to return
	imageReaders := []bytes.Reader{}
	for _, imageBlock := range response.Output.Data.Images {
		imageReader, err := downloadImageAsReader(imageBlock.Filename)
		if err != nil {
			return []bytes.Reader{}, ErrFetchImages
		}

		imageReaders = append(imageReaders, *imageReader)
	}

	return imageReaders, nil
}

var (
	ErrDownloadFailed = errors.New("Failed to GET file")
	ErrReadFailed     = errors.New("Failed to read response bytes")
)

func downloadImageAsReader(filepath string) (*bytes.Reader, error) {
	url := fmt.Sprintf(
		"http://%s/file=%s",
		viper.GetString("stable_diffusion.host"),
		filepath,
	)

	res, err := http.Get(url)
	if err != nil {
		return nil, ErrDownloadFailed
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, ErrReadFailed
	}

	res.Body.Close()

	return bytes.NewReader(data), nil
}
