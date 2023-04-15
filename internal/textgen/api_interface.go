package textgen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/M-Ro/aurora-ai/api"
	"github.com/M-Ro/aurora-ai/config"
	"github.com/M-Ro/aurora-ai/internal/helpers"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type InferenceUpdateFunc func(string)
type InferenceCompleteFunc func()

type PacketMode int

const (
	Prepare PacketMode = iota
    Inference	
)

func RunInference(
	query string,
	onUpdate InferenceUpdateFunc,
	onComplete InferenceCompleteFunc,
) error {
	s := GetSession()
	apiConn := newAPIConnection()

	// First we need to open the connection to send the prep statement
	err := apiConn.connect()
	if err != nil {
		return err
	}

	mode := Prepare
	err = runSocketHandler(apiConn, query, s, mode, onUpdate, onComplete)
    if err != nil {
        logrus.Error("error: ", err)
        return err
    }

	// The first connection will be terminated by the server, run again to send the actual
	// inference statement
	err = apiConn.connect()
	if err != nil {
		return err
	}

	mode = Inference 
	err = runSocketHandler(apiConn, query, s, mode, onUpdate, onComplete)

	return err
}

func runSocketHandler(
	conn *APIConnection,
	query string,
	s *Session,
	mode PacketMode,
	onUpdate InferenceUpdateFunc,
	onComplete InferenceCompleteFunc,
) error {

	sendQueue := make(chan string)
	done := make(chan struct{})
	go func() {
		defer close(done)

		for {
            logrus.Info("Calling ReadMessage() to block")
			_, message, err := conn.Ws.ReadMessage()
			if err != nil {
				logrus.Error("ws read: ", err)
				return
			}

			logrus.Info("ws recv: ", string(message))

			respPacket := api.GradioResponsePacket{}
			err = json.Unmarshal(message, &respPacket)
			if err != nil {
				logrus.Error("Failed to unmarshal response: ", err)
				return
			}

			switch respPacket.Message {
			case api.MsgSendHash:
				onServerRequestHash(s, mode, sendQueue)
			case api.MsgSendData:
				onServerRequestData(s, mode, query, sendQueue)
			case api.MsgProcessCompleted:
				onServerProcessComplete(s, &mode, sendQueue)
				onComplete()
				return
			case api.MsgProcessGenerating:
				dataStr, err := onServerProcessGenerating(&respPacket)
				if err != nil {
					logrus.Error("Failed handling ProcessGenerating packet")
					return
				}
				onUpdate(dataStr)
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

func onServerRequestHash(session *Session, mode PacketMode, sendQueue chan string) {
	logrus.Info("Sending auth hash: ", session.SessionHash)

	fnIndex, err := getFnIndex(mode)
	bytes, err := json.Marshal(api.SendHashRequest{
		SessionHash: session.SessionHash,
		FnIndex:     fnIndex,
	})

	if err != nil {
        logrus.Error("onServerRequestHash: Failed marshalling")
		return
	}

	sendQueue <- string(bytes)
}

func onServerRequestData(session *Session, mode PacketMode, query string, sendQueue chan string) {
	logrus.Info("Sending data")

	fnIndex, err := getFnIndex(mode)
    if err != nil {
        logrus.Error("well shit, err isnt nil", err)
    }

    // this is stupid, do a custom marshaller for this bullshit later
    if mode == Prepare {
        sillyString := fmt.Sprintf(
            "{\"fn_index\":8, \"data\":[1512,-1,1.99,0.18,30,1,1.15,1,0,0,true,0,1,1,false,true,\"\\\"\\\\n### Human:\\\", \\\"\\\\n### Assistant:\\\"\"],\"event_data\":null,\"session_hash\":\"%s\"}",
            session.SessionHash,
        )

        sendQueue <- sillyString
    } else if mode == Inference {
        bytes, err := json.Marshal(api.SendInferenceDataRequest{
            SessionHash: session.SessionHash,
            FnIndex:     fnIndex,
            Data:        getData(mode, &query),
        })

        if err != nil {
            logrus.Error("onServerRequestData: Failed to marshal")
            return
        }

	    sendQueue <- string(bytes)
    }
}

func onServerProcessComplete(session *Session, mode *PacketMode, sendQueue chan string) {
	if *mode == Prepare {
		*mode = Inference 
		return
	}

}

func onServerProcessGenerating(packet *api.GradioResponsePacket) (string, error) {
    return getBotStringFromResponse(packet.Output.Data[0])
}

// getBotStringFromResponse isolates & extracts the actual bot response from the output
func getBotStringFromResponse(response string) (string, error) {
    // Extract the actual response
    botToken := viper.GetString("llm.identifier_b")

    // We need to remove the last string from the human token
    // why? because the api is inconsistent and requires a suffix colon to inference without
    // going schizo, but at termination doesn't bother to produce a suffix colon itself.
    humanToken := viper.GetString("llm.identifier_p")
    humanToken = strings.TrimRight(humanToken, ":")

    lB := strings.LastIndex(response, botToken)
    lH := strings.LastIndex(response, humanToken)

    // If lH > lB, the bot has re-prompted the user, so fetch the string upto that point
    if lH > lB {
        a := helpers.Substr(
            response,
            lB + len(botToken),
            lH - (lB + len(botToken)),
        )

        return a, nil
    }

    return helpers.Substr(
        response,
        lB + len(botToken),
        len(response) - (lB + len(botToken)),
    ), nil
}

func getFnIndex(mode PacketMode) (uint32, error) {
	switch mode {
	case Prepare:
		return config.FnSendNoFuckingIdea8, nil
	case Inference:
		return config.FnSendNoFuckingIdea9, nil
	}

	//FIXME err
	return 0, nil
}

func getData(mode PacketMode, query *string) []*string {
	switch mode {
	//case Prepare:
		//return []*string{query}
	case Inference:
        context := viper.GetString("llm.context")
        botToken := viper.GetString("llm.identifier_b")
        humanToken := viper.GetString("llm.identifier_p")

        output := fmt.Sprintf("%s\n%s \n%s\n%s", context, humanToken, *query, botToken) 
		return []*string{
            &output,
			nil,
		}
	default:
		return []*string{}
	}
}
