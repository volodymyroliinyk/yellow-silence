package capture

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	portalBus       = "org.freedesktop.portal.Desktop"
	portalPath      = dbus.ObjectPath("/org/freedesktop/portal/desktop")
	requestResponse = "org.freedesktop.portal.Request.Response"
	requestClose    = "org.freedesktop.portal.Request.Close"
	sessionClose    = "org.freedesktop.portal.Session.Close"
	portalFrameRate = 7
)

type portalCapturer struct {
	gstPath string

	mu          sync.Mutex
	initialized bool
	initErr     error
	conn        *dbus.Conn
	signals     chan *dbus.Signal
	session     dbus.ObjectPath
	cmd         *exec.Cmd
	stdout      io.ReadCloser
	width       int
	height      int
	frameBytes  int
}

func newPortalCapturer(gstPath string) *portalCapturer {
	return &portalCapturer{gstPath: gstPath}
}

func (p *portalCapturer) Capture(ctx context.Context) (image.Image, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.initialized {
		p.initialized = true
		p.initErr = p.initialize(ctx)
	}
	if p.initErr != nil {
		return nil, p.initErr
	}

	pixels := make([]byte, p.frameBytes)
	if _, err := io.ReadFull(p.stdout, pixels); err != nil {
		clear(pixels)
		return nil, fmt.Errorf("read Wayland frame: %w", err)
	}
	return &image.RGBA{
		Pix:    pixels,
		Stride: p.width * 4,
		Rect:   image.Rect(0, 0, p.width, p.height),
	}, nil
}

func (p *portalCapturer) initialize(ctx context.Context) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("connect to desktop portal: %w", err)
	}
	p.conn = conn
	p.signals = make(chan *dbus.Signal, 16)
	conn.Signal(p.signals)
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.portal.Request"),
		dbus.WithMatchMember("Response"),
	); err != nil {
		return fmt.Errorf("listen for desktop portal response: %w", err)
	}

	sessionToken, err := randomToken("yellow_silence_session_")
	if err != nil {
		return fmt.Errorf("create portal session token: %w", err)
	}
	createToken, err := randomToken("yellow_silence_create_")
	if err != nil {
		return fmt.Errorf("create portal request token: %w", err)
	}
	results, err := p.callRequest(ctx, "org.freedesktop.portal.ScreenCast.CreateSession", map[string]dbus.Variant{
		"handle_token":         dbus.MakeVariant(createToken),
		"session_handle_token": dbus.MakeVariant(sessionToken),
	})
	if err != nil {
		return err
	}
	value, ok := results["session_handle"]
	if !ok {
		return errors.New("desktop portal did not return a session")
	}
	switch path := value.Value().(type) {
	case dbus.ObjectPath:
		p.session = path
	case string:
		p.session = dbus.ObjectPath(path)
	default:
		return errors.New("desktop portal returned an invalid session")
	}
	if !p.session.IsValid() {
		return errors.New("desktop portal returned an invalid session path")
	}

	selectToken, err := randomToken("yellow_silence_select_")
	if err != nil {
		return fmt.Errorf("create portal request token: %w", err)
	}
	_, err = p.callRequest(ctx, "org.freedesktop.portal.ScreenCast.SelectSources", p.session, map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant(selectToken),
		"types":        dbus.MakeVariant(uint32(1)),
		"multiple":     dbus.MakeVariant(false),
		"cursor_mode":  dbus.MakeVariant(uint32(1)),
		"persist_mode": dbus.MakeVariant(uint32(1)),
	})
	if err != nil {
		return err
	}

	startToken, err := randomToken("yellow_silence_start_")
	if err != nil {
		return fmt.Errorf("create portal request token: %w", err)
	}
	results, err = p.callRequest(ctx, "org.freedesktop.portal.ScreenCast.Start", p.session, "", map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant(startToken),
	})
	if err != nil {
		return err
	}
	stream, err := firstStream(results)
	if err != nil {
		return err
	}
	if !safeDimensions(stream.width, stream.height) || stream.width > maxCaptureBytes/4/stream.height {
		return errors.New("Wayland frame dimensions exceed safety limit")
	}
	p.width = stream.width
	p.height = stream.height
	p.frameBytes = stream.width * stream.height * 4

	var remote dbus.UnixFD
	call := conn.Object(portalBus, portalPath).CallWithContext(ctx,
		"org.freedesktop.portal.ScreenCast.OpenPipeWireRemote", 0, p.session, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("open portal PipeWire remote: %w", call.Err)
	}
	if err := call.Store(&remote); err != nil {
		return fmt.Errorf("read portal PipeWire remote: %w", err)
	}
	remoteFile := os.NewFile(uintptr(remote), "portal-pipewire")
	if remoteFile == nil {
		return errors.New("desktop portal returned an invalid PipeWire descriptor")
	}
	defer remoteFile.Close()

	args := []string{
		"-q", "pipewiresrc", "fd=3", "path=" + strconv.FormatUint(uint64(stream.nodeID), 10),
		"do-timestamp=true", "!", "videorate", "drop-only=true", "max-rate=" + strconv.Itoa(portalFrameRate), "!",
		"videoconvert", "!", fmt.Sprintf("video/x-raw,format=RGBA,width=%d,height=%d,framerate=%d/1", stream.width, stream.height, portalFrameRate), "!",
		"fdsink", "fd=1", "sync=false",
	}
	cmd := exec.CommandContext(ctx, p.gstPath, args...)
	cmd.ExtraFiles = []*os.File{remoteFile}
	cmd.Stderr = io.Discard
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create Wayland frame pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		stdout.Close()
		return fmt.Errorf("start Wayland frame receiver: %w", err)
	}
	p.cmd = cmd
	p.stdout = stdout
	return nil
}

func (p *portalCapturer) callRequest(ctx context.Context, method string, args ...any) (map[string]dbus.Variant, error) {
	call := p.conn.Object(portalBus, portalPath).CallWithContext(ctx, method, 0, args...)
	if call.Err != nil {
		return nil, fmt.Errorf("desktop portal request failed: %w", call.Err)
	}
	var requestPath dbus.ObjectPath
	if err := call.Store(&requestPath); err != nil {
		return nil, fmt.Errorf("read desktop portal request: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			_ = p.conn.Object(portalBus, requestPath).Call(requestClose, 0).Err
			return nil, ctx.Err()
		case signal := <-p.signals:
			if signal == nil || signal.Path != requestPath || signal.Name != requestResponse {
				continue
			}
			if len(signal.Body) != 2 {
				return nil, errors.New("desktop portal returned a malformed response")
			}
			status, ok := signal.Body[0].(uint32)
			if !ok {
				return nil, errors.New("desktop portal returned an invalid status")
			}
			results, ok := signal.Body[1].(map[string]dbus.Variant)
			if !ok {
				return nil, errors.New("desktop portal returned invalid results")
			}
			switch status {
			case 0:
				return results, nil
			case 1:
				return nil, errors.New("screen sharing was cancelled by the user")
			default:
				return nil, errors.New("desktop portal denied screen sharing")
			}
		}
	}
}

func (p *portalCapturer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stdout != nil {
		_ = p.stdout.Close()
	}
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
	}
	if p.conn != nil {
		if p.session.IsValid() {
			_ = p.conn.Object(portalBus, p.session).Call(sessionClose, 0).Err
		}
		if p.signals != nil {
			p.conn.RemoveSignal(p.signals)
		}
		return p.conn.Close()
	}
	return nil
}

func randomToken(prefix string) (string, error) {
	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(random[:]), nil
}

type portalStream struct {
	nodeID uint32
	width  int
	height int
}

func firstStream(results map[string]dbus.Variant) (portalStream, error) {
	value, ok := results["streams"]
	if !ok {
		return portalStream{}, errors.New("desktop portal returned no screen stream")
	}
	var first []any
	switch streams := value.Value().(type) {
	case [][]any:
		if len(streams) > 0 {
			first = streams[0]
		}
	case []any:
		if len(streams) > 0 {
			first, _ = streams[0].([]any)
		}
	}
	if len(first) < 1 {
		return portalStream{}, errors.New("desktop portal returned an invalid screen stream")
	}
	nodeID, ok := first[0].(uint32)
	if !ok || nodeID == 0 {
		return portalStream{}, errors.New("desktop portal returned an invalid PipeWire node")
	}
	if len(first) < 2 {
		return portalStream{}, errors.New("desktop portal returned no screen dimensions")
	}
	properties, ok := first[1].(map[string]dbus.Variant)
	if !ok {
		return portalStream{}, errors.New("desktop portal returned invalid screen properties")
	}
	size, ok := properties["size"]
	if !ok {
		return portalStream{}, errors.New("desktop portal returned no screen dimensions")
	}
	width, height, ok := int32Pair(size.Value())
	if !ok || width < 1 || height < 1 {
		return portalStream{}, errors.New("desktop portal returned invalid screen dimensions")
	}
	return portalStream{nodeID: nodeID, width: int(width), height: int(height)}, nil
}

func int32Pair(value any) (int32, int32, bool) {
	switch pair := value.(type) {
	case []int32:
		if len(pair) == 2 {
			return pair[0], pair[1], true
		}
	case []any:
		if len(pair) == 2 {
			left, leftOK := pair[0].(int32)
			right, rightOK := pair[1].(int32)
			return left, right, leftOK && rightOK
		}
	}
	return 0, 0, false
}
