package kit

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func tarOf(files []stored, manifest []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := tar.NewWriter(&buf)
	if err := writeTar(writer, "manifest.json", manifest); err != nil {
		return nil, err
	}
	for _, file := range files {
		if err := writeTar(writer, file.Rel, file.Data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeTar(writer *tar.Writer, rel string, body []byte) error {
	clean, err := safeSlash(rel)
	if err != nil {
		return err
	}
	header := &tar.Header{
		Name:     clean,
		Mode:     0o600,
		Size:     int64(len(body)),
		Typeflag: tar.TypeReg,
	}
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	_, err = writer.Write(body)
	return err
}

func untar(body []byte, dest string) error {
	reader := tar.NewReader(bytes.NewReader(body))
	count := 0
	var total int64
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		count++
		if count > maxCount {
			return ErrTooBig
		}
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			return ErrEscape
		}
		name := strings.TrimSuffix(header.Name, "/")
		if name == "." || name == "" {
			continue
		}
		target, err := joinUnder(dest, name)
		if err != nil {
			return err
		}
		if header.FileInfo().IsDir() || strings.HasSuffix(header.Name, "/") {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		if header.Size < 0 || header.Size > maxFile || total+header.Size > maxTotal {
			return ErrTooBig
		}
		total += header.Size
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, io.LimitReader(reader, header.Size)); err != nil {
			_ = file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
}

// Verify checks hashes. An encrypted kit without a passphrase returns ErrPass and does not write anywhere.
func Verify(path string, passphrase []byte, work int) (Report, error) {
	dir, cleanup, report, err := openKit(path, passphrase)
	defer cleanup()
	if err != nil {
		return report, err
	}
	if dir == "" {
		return report, errors.New("kit")
	}
	report.Verified = true
	return report, nil
}

func openKit(path string, passphrase []byte) (string, func(), Report, error) {
	noop := func() {}
	info, err := os.Lstat(path)
	if err != nil {
		return "", noop, Report{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", noop, Report{}, ErrEscape
	}
	if info.IsDir() {
		manifest, err := readManifest(path)
		if err != nil {
			return "", noop, Report{}, err
		}
		return path, noop, Report{Manifest: manifest, Label: manifest.Label}, nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", noop, Report{}, err
	}
	if int64(len(body)) > maxTotal {
		return "", noop, Report{}, ErrTooBig
	}
	label, date, payloadErr := headerOnly(body)
	report := Report{Encrypted: true, Label: label}
	if payloadErr != nil && !bytes.HasPrefix(body, []byte("WOWKIT1\n")) {
		return "", noop, Report{}, payloadErr
	}
	if len(bytes.TrimSpace(passphrase)) == 0 {
		report.Manifest.CreatedAt = date
		report.Manifest.Label = label
		return "", noop, report, ErrPass
	}
	plain, label, date, err := decryptPayload(body, passphrase)
	report.Label = label
	report.Manifest.CreatedAt = date
	report.Manifest.Label = label
	if err != nil {
		return "", noop, report, err
	}
	tmp, err := os.MkdirTemp("", "bornal-kit-")
	if err != nil {
		return "", noop, report, err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }
	if err := untar(plain, tmp); err != nil {
		cleanup()
		return "", noop, report, err
	}
	manifest, err := readManifest(tmp)
	if err != nil {
		cleanup()
		return "", noop, report, err
	}
	report.Manifest = manifest
	report.Label = label
	return tmp, cleanup, report, nil
}

func headerOnly(body []byte) (string, string, error) {
	label, date, _, err := splitHeader(body)
	return label, date, err
}
