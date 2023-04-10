package textgen

import (
	"encoding/json"

	"github.com/M-Ro/aurora-ai/api"
	"github.com/M-Ro/aurora-ai/config"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type InferenceUpdateFunc func(string)
type InferenceCompleteFunc func()

type PacketMode int
const (
    Prepare PacketMode = iota
    Instruct
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

    // The first connection will be terminated by the server, run again to send the actual
    // inference statement
    err = apiConn.connect()
    if err != nil {
        return err
    }

    mode = Instruct
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
            _, message, err := conn.Ws.ReadMessage()
            if err != nil {
                logrus.Error("ws read: ", err)
                return
            }

            //logrus.Info("ws recv: ", string(message))

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
        FnIndex: fnIndex, 
    })

    if err != nil {
        logrus.Error("fek")
        return
    }

    sendQueue <- string(bytes)
}

func onServerRequestData(session *Session, mode PacketMode, query string, sendQueue chan string) {
    logrus.Info("Sending data")

    fnIndex, err := getFnIndex(mode)
    bytes, err := json.Marshal(api.SendDataRequest{
        SessionHash: session.SessionHash,
        FnIndex: fnIndex, 
        Data: getData(mode, &query),
    })

    if err != nil {
        logrus.Error("fek")
        return
    }

    sendQueue <- string(bytes)
}

func onServerProcessComplete(session *Session, mode *PacketMode, sendQueue chan string) {
    if *mode == Prepare {
        *mode = Instruct
        return
    }

}

func onServerProcessGenerating(packet *api.GradioResponsePacket) (string, error) {
    return packet.Output.Data[0], nil
}

func getFnIndex(mode PacketMode) (uint32, error) {
    switch mode {
        case Prepare:
            return config.FnSendPrepQuery, nil
        case Instruct:
            return config.FnSendInstructQuery, nil
    }

    //FIXME err
    return 0, nil
}

func getData(mode PacketMode, query *string) []string { 
        switch mode {
        case Prepare: return []string{*query}
        case Instruct: return []string{
            "", // SHOULD BE NIL
            "", // SHOULD BE NIL
            "### Human:",
            "### Assistant:",
            "Below is an instruction that describes a task. Write a response that appropriately completes the request.\n\n",
            "instruct",
            "",
        }
        default: return []string{}
    }
}
