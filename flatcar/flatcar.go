package flatcar

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/tosuke/hokuchi"
	"golang.org/x/sync/errgroup"
)

type Fetcher struct {
	http *http.Client
	sema chan struct{}

	// cached
	_signKeyring *crypto.KeyRing
}
type Option struct {
	RequestConcurrency int
}

func New(option Option) *Fetcher {
	var sema chan struct{}
	if option.RequestConcurrency > 0 {
		sema = make(chan struct{}, option.RequestConcurrency)
	}

	return &Fetcher{
		http: http.DefaultClient,
		sema: sema,
	}
}

func (f *Fetcher) ResolveKey(ctx context.Context, channel, arch, version string) (Key, error) {
	k := Key{
		channel: channel,
		arch:    hokuchi.NormalizeArch(arch),
		version: version,
	}

	if !isChannelValid(k.channel) {
		return Key{}, errors.New("invalid channel")
	}
	if !isArchValid(k.arch) {
		return Key{}, errors.New("invalid arch")
	}
	if k.version == "current" {
		currentVer, err := f.fetchVersion(ctx, k)
		if err != nil {
			return Key{}, fmt.Errorf("failed to fetch current version: %w", err)
		}
		k.version = currentVer
	}
	if !isVersionValid(k.version) {
		return Key{}, errors.New("invalid version")
	}

	return k, nil
}

func (f *Fetcher) FetchKernel(ctx context.Context, w io.Writer, key Key) error {
	return f.fetchKernel(ctx, w, key)
}

func (f *Fetcher) fetchVersion(ctx context.Context, key Key) (string, error) {
	var versionBuf bytes.Buffer
	var versionSigBuf bytes.Buffer

	eg, egctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := f.fetchData(egctx, &versionBuf, key, "/version.txt", 2048); err != nil {
			return fmt.Errorf("failed to fetch version data: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := f.fetchData(egctx, &versionSigBuf, key, "/version.txt.sig", 2048); err != nil {
			return fmt.Errorf("failed to fetch version signature: %w", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return "", err
	}

	message := crypto.NewPlainMessage(versionBuf.Bytes())
	signature := crypto.NewPGPSignature(versionSigBuf.Bytes())

	signKeyring, err := f.getSignKeyring()
	if err != nil {
		return "", fmt.Errorf("failed to get flatcar sign keyring: %w", err)
	}
	if err := signKeyring.VerifyDetached(message, signature, crypto.GetUnixTime()); err != nil {
		return "", fmt.Errorf("failed to verify flatcar version: %w", err)
	}

	sc := bufio.NewScanner(&versionBuf)
	for sc.Scan() {
		line := sc.Text()
		if after, ok := strings.CutPrefix(line, "FLATCAR_VERSION="); ok {
			version := strings.TrimSpace(after)
			return version, nil
		}
	}
	return "", fmt.Errorf("failed to find version")
}

func (f *Fetcher) fetchKernel(ctx context.Context, w io.Writer, key Key) error {
	eg, ectx := errgroup.WithContext(ctx)
	pr, pw := io.Pipe()

	eg.Go(func() error {
		writer := io.MultiWriter(w, pw)
		defer pw.Close()
		if err := f.fetchData(ectx, writer, key, "/flatcar_production_pxe.vmlinuz", 0); err != nil {
			return fmt.Errorf("failed to fetch kernel: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		signKeyring, err := f.getSignKeyring()
		if err != nil {
			return fmt.Errorf("failed to get flatcar sign keying: %w", err)
		}

		var kernelSigBuf bytes.Buffer
		if err := f.fetchData(ctx, &kernelSigBuf, key, "/flatcar_production_pxe.vmlinuz.sig", 2048); err != nil {
			return fmt.Errorf("failed to fetch kernel signature: %w", err)
		}
		signature := crypto.NewPGPSignature(kernelSigBuf.Bytes())

		if err := signKeyring.VerifyDetachedStream(pr, signature, crypto.GetUnixTime()); err != nil {
			return fmt.Errorf("failed to verify kernel: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (f *Fetcher) fetchData(ctx context.Context, w io.Writer, key Key, subpath string, limit int64) error {
	url := key.baseURL() + subpath
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for flatcar repository: %w", err)
	}

	if f.sema != nil {
		f.sema <- struct{}{}
		defer func() {
			<-f.sema
		}()
	}

	slog.DebugContext(ctx, "request", slog.String("url", url))
	resp, err := f.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request flatcar data: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	slog.DebugContext(ctx, "processing response", slog.String("url", url))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status from flatcar: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var r io.Reader = resp.Body
	if limit > 0 {
		r = io.LimitReader(r, limit)
	}
	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("failed to copy flatcar response: %w", err)
	}
	return nil
}

//go:embed Flatcar_Image_Signing_key.asc
var flatcarGPGKey string

func (f *Fetcher) getSignKeyring() (*crypto.KeyRing, error) {
	if f._signKeyring != nil {
		return f._signKeyring, nil
	}

	pubKeyObj, err := crypto.NewKeyFromArmored(flatcarGPGKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flatcar key: %w", err)
	}
	signKeyring, err := crypto.NewKeyRing(pubKeyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to create flatcar keyring: %w", err)
	}

	f._signKeyring = signKeyring
	return signKeyring, nil
}
