package protonostale

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
	"github.com/sirupsen/logrus"
)

var (
	HandoffOpCode = "NoS0575"
)

func NewAuthMessageReader(r io.Reader) *AuthMessageReader {
	return &AuthMessageReader{
		MessageReader: MessageReader{
			r: NewFieldReader(r),
		},
	}
}

type AuthMessageReader struct {
	MessageReader
}

func (r *AuthMessageReader) ReadRequestHandoffMessage() (*eventing.RequestHandoffMessage, error) {
	logrus.Debugf("reading request handoff message")
	_, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading identifier")
	identifier, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading password")
	pwd, err := r.readPassword()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading client version")
	clientVersion, err := r.readVersion()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading op code")
	return &eventing.RequestHandoffMessage{
		Identifier:    identifier,
		Password:      pwd,
		ClientVersion: clientVersion,
	}, nil
}

func (r *AuthMessageReader) readPassword() (string, error) {
	pwd, err := r.r.ReadField()
	if err != nil {
		return "", err
	}
	if len(pwd)%2 == 0 {
		pwd = pwd[3:]
	} else {
		pwd = pwd[4:]
	}
	chunks := bytesChunkEvery(pwd, 2)
	pwd = make([]byte, 0, len(chunks))
	for i := 0; i < len(chunks); i++ {
		pwd = append(pwd, chunks[i][0])
	}
	chunks = bytesChunkEvery(pwd, 2)
	var result []byte
	for _, chunk := range chunks {
		value, err := strconv.ParseInt(string(chunk), 16, 64)
		if err != nil {
			return "", err
		}
		result = append(result, byte(value))
	}

	return string(result), nil
}

var versionRegex = regexp.MustCompile(`.+\v(\d+).(\d+).(\d+).(\d+)\n`)

func (r *AuthMessageReader) readVersion() (string, error) {
	version, err := r.r.ReadField()
	if err != nil {
		return "", err
	}
	logrus.Debugf("version: %v", version)
	matches := versionRegex.FindAllSubmatch(version, 1)
	logrus.Debugf("version matches: %v", matches)
	if len(matches) != 1 || len(matches[0]) != 5 {
		return "", fmt.Errorf("invalid version format: %s", string(version))
	}
	major := matches[0][1]
	minor := matches[0][2]
	patch := matches[0][3]
	build := matches[0][4]
	return fmt.Sprintf("%s.%s.%s+%s", major, minor, patch, build), nil
}

func NewAuthReader(r *bufio.Reader) *AuthReader {
	return &AuthReader{
		r: simplesubtitution.NewReader(r),
	}
}

type AuthReader struct {
	r *simplesubtitution.Reader
}

func (r *AuthReader) ReadMessage() (*AuthMessageReader, error) {
	buff, err := r.r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("auth: read %s messages", string(buff))
	return NewAuthMessageReader(bytes.NewReader(buff)), nil
}

func NewAuthWriter(w *bufio.Writer) *AuthWriter {
	return &AuthWriter{
		Writer: Writer{
			w: bufio.NewWriter(simplesubtitution.NewWriter(w)),
		},
	}
}

type AuthWriter struct {
	Writer
}

func (w *AuthWriter) WriteProposeHandoffMessage(
	msg *eventing.ProposeHandoffMessage,
) error {
	_, err := fmt.Fprintf(w.w, "%s %d", "NsTeST", msg.Key)
	if err != nil {
		return err
	}
	for _, g := range msg.Gateways {
		if err := w.WriteGatewayMessage(g); err != nil {
			return err
		}
		if err := w.w.WriteByte(' '); err != nil {
			return err
		}
	}
	if _, err = w.w.WriteString("-1 -1 -1 10000 10000 1"); err != nil {
		return err
	}
	return w.w.Flush()
}

func (w *AuthWriter) WriteGatewayMessage(
	msg *eventing.GatewayMessage,
) error {
	_, err := fmt.Fprintf(
		w.w,
		"%s %s %d %.f %d %d %s",
		msg.Host,
		msg.Port,
		msg.Population,
		math.Round(float64(msg.Population)/float64(msg.Capacity)*20)+1,
		msg.WorldId,
		msg.ChannelId,
		msg.WorldName,
	)
	return err
}
