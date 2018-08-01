package sync

import (
	"./notFound"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log/syslog"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/exec"
	"time"
)

var (
	Root     = flag.String("root", "/", "The path of the root")
	Contents = flag.String("contents_root", "/var/www/html/", "The path of the contents root")
	EndPoint = flag.String("sync_endpoint", "/sns_notify", "The path of the endpoint")
	NotFound = flag.String("nfpage", "404.html", "The filename of the custome 404 page")
	ListenIP = flag.String("listenIP", "127.0.0.1:9000", "The Listen IP address and port number")
	Log      *syslog.Writer
)

const (
	Notification             = "Notification"
	SubscriptionConfirmation = "SubscriptionConfirmation"
)

type Message struct {
	Type             string
	MessageId        string
	Token            string
	TopicArn         string
	Subject          string
	Message          string
	Timestamp        time.Time
	SignatureVersion string
	Signature        string
	SigningCertURL   string
	SubscribeURL     string
	UnsubscribeURL   string
}

type AppServer struct {
	Protocol string
}

func newMessageFromJSON(js []byte) (*Message, error) {
	snsMsg := &Message{}
	err := json.Unmarshal(bytes.Trim(js, "\x00"), &snsMsg)
	if err != nil {
		return nil, err
	}

	if loc, err := time.LoadLocation("Local"); err != nil {
		return snsMsg, err
	} else {
		snsMsg.Timestamp = snsMsg.Timestamp.In(loc)
		return snsMsg, nil
	}
}

func (sm *Message) String() string {
	return fmt.Sprintf("%s [%s] %s", sm.Timestamp.Format(time.RFC3339), sm.Subject, sm.Message)
}

func (sm *Message) confirmSubscription() error {
	_, err := http.Get(sm.SubscribeURL)
	if err != nil {
		return err
	}
	return nil
}

func outErr(w http.ResponseWriter, err error) bool {
	if err != nil {
		Log.Err(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	} else {
		return true
	}
}

func sync(sm *Message) (err error) {
	return exec.Command("sudo", "systemctl", "start", "getcontents.service").Run()
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	if body, err := ioutil.ReadAll(r.Body); !outErr(w, err) {
		return
	} else {
		if snsMsg, err := newMessageFromJSON(body); !outErr(w, err) {
			return
		} else {
			switch snsMsg.Type {
			case Notification:
				Log.Info(fmt.Sprintf("[Vars: %#v] [SNS: %#v]", vars, snsMsg))
				outErr(w, sync(snsMsg))
			case SubscriptionConfirmation:
				Log.Info(fmt.Sprintf("[Vars: %#v] [SNS: %#v]", vars, snsMsg))
				outErr(w, snsMsg.confirmSubscription())
			}
		}
	}
}

func (app *AppServer) Serve() {
	defer Log.Close()
	if l, err := net.Listen(app.Protocol, *ListenIP); err != nil {
		Log.Err(err.Error())
		os.Exit(1)
	} else {
		defer fcgi.Serve(l, nil)
		router := mux.NewRouter()
		router.HandleFunc(*EndPoint, hookHandler).Methods("POST")
		http.Handle(*EndPoint, router)
		http.Handle("/", notFound.Handle404(http.StripPrefix(*Root, http.FileServer(http.Dir(*Contents))), func(w http.ResponseWriter, r *http.Request) bool {
			notFound.TryRead404(w, *Contents + *NotFound)
			return true
		}))
	}
}

func NewAppServer(protocol string) (l *AppServer, err error) {
	l = new(AppServer)
	l.Protocol = protocol
	Log, err = syslog.New(syslog.LOG_INFO, os.Args[0])
	return
}
