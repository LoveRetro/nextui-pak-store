package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/skip2/go-qrcode"
)

func GetPlatform() string {
	raw := strings.ToLower(os.Getenv("PLATFORM"))

	switch models.Platform(raw) {
	case models.TG5040:
		return string(models.TG5040)
	case models.TG5050:
		return string(models.TG5050)
	default:
		return string(models.TG5040)
	}
}

func GetSDRoot() string {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		return os.Getenv("SD_ROOT")
	}

	return models.SDRoot
}

func GetUserDataDir() string {
	platform := GetPlatform()
	return filepath.Join(GetSDRoot(), models.UserdataDir, platform, models.PakStoreUserDataDir)
}

func GetToolRoot() string {
	return filepath.Join(GetSDRoot(), models.ToolDir, GetPlatform())
}

func GetEmulatorRoot() string {
	return filepath.Join(GetSDRoot(), models.EmulatorDir, GetPlatform())
}

func LogStandardFatal(msg string, err error) {
	log.SetOutput(os.Stderr)
	log.Fatalf("%s: %v", msg, err)
}

func FetchStorefront() (models.Storefront, error) {
	logger := gabagool.GetLogger()

	var data []byte
	var err error

	if override := os.Getenv("STOREFRONT_OVERRIDE"); override != "" {
		data, err = fetch(override)
		if err != nil {
			return models.Storefront{}, err
		}
	} else if os.Getenv("ENVIRONMENT") == "DEV" {
		data, err = os.ReadFile("storefront.json")
		if err != nil {
			return models.Storefront{}, fmt.Errorf("failed to read local storefront.json: %w", err)
		}
	} else {
		data, err = fetch(models.StorefrontJsonURL)
		if err != nil {
			return models.Storefront{}, err
		}
	}

	var sf models.Storefront
	if err := json.Unmarshal(data, &sf); err != nil {
		return models.Storefront{}, err
	}

	for i, p := range sf.Paks {
		if filepath.Ext(p.ReleaseFilename) == ".pakz" {
			sf.Paks[i].IsPakZ = true
		}
	}

	logger.Info("Fetched storefront", "name", sf.Name)

	return sf, nil
}

func ParseJSONFile(filePath string, out *models.Pak) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func DownloadPakArchive(pak models.Pak) (tempFile string, completed bool, error error) {
	logger := gabagool.GetLogger()

	releasesStub := fmt.Sprintf("/releases/download/%s/", pak.Version)
	dl := pak.RepoURL + releasesStub + pak.ReleaseFilename
	tmp := filepath.Join("/tmp", pak.ReleaseFilename)

	message := fmt.Sprintf("Downloading %s %s...", pak.StorefrontName, pak.Version)

	res, err := gabagool.DownloadManager([]gabagool.Download{{
		URL:         dl,
		Location:    tmp,
		DisplayName: message,
	}}, make(map[string]string), gabagool.DownloadManagerOptions{AutoContinueOnComplete: true})

	if err != nil {
		// Check if it was cancelled
		if errors.Is(err, gabagool.ErrCancelled) {
			return "", false, nil
		}
		logger.Error("Error downloading", "error", err)
		return "", false, err
	}

	// Check for failed downloads
	if len(res.Failed) > 0 {
		err = res.Failed[0].Error
		logger.Error("Error downloading", "error", err)
		return "", false, err
	}

	return tmp, true, nil
}

func RunScript(script models.Script, scriptName string) error {
	logger := gabagool.GetLogger()

	if script.Path == "" {
		logger.Info("No script to run")
		return nil
	}

	_, err := gabagool.ProcessMessage(fmt.Sprintf("%s %s %s...", "Running", scriptName, "Script"), gabagool.ProcessMessageOptions{}, func() (interface{}, error) {
		logger.Info("Running script",
			"path", script.Path,
			"args", script.Args)

		cmd := exec.Command(script.Path, script.Args...)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			logger.Error("Failed to execute script",
				"error", err,
				"path", script.Path,
				"args", script.Args,
				"stderr", stderr.String())
			return nil, fmt.Errorf("failed to execute script %s: %w", script.Path, err)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			logger.Error("Script returned non-zero exit code",
				"path", script.Path,
				"args", script.Args,
				"exitCode", cmd.ProcessState.ExitCode(),
				"stderr", stderr.String())
			return nil, fmt.Errorf("script %s exited with code %d: %s",
				script.Path, cmd.ProcessState.ExitCode(), stderr.String())
		}

		logger.Info("Script executed successfully",
			"path", script.Path,
			"args", script.Args,
			"stdout", stdout.String())

		return nil, nil
	})

	return err
}

func UnzipPakArchive(pak models.Pak, tmp string) error {
	logger := gabagool.GetLogger()

	pakDestination := ""

	if pak.IsPakZ {
		pakDestination = GetSDRoot()
	} else if pak.PakType == models.PakTypes.TOOL {
		pakDestination = filepath.Join(GetToolRoot(), pak.Name+".pak")
	} else if pak.PakType == models.PakTypes.EMU {
		pakDestination = filepath.Join(GetEmulatorRoot(), pak.Name+".pak")
	}

	_, err := gabagool.ProcessMessage(fmt.Sprintf("%s %s...", "Unzipping", pak.StorefrontName), gabagool.ProcessMessageOptions{}, func() (interface{}, error) {
		err := Unzip(tmp, pakDestination, pak, false)
		if err != nil {
			return nil, err
		}

		time.Sleep(1 * time.Second)

		return nil, nil
	})

	if err != nil {
		gabagool.ProcessMessage(fmt.Sprintf("Unable to unzip %s", pak.StorefrontName), gabagool.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return nil, nil
		})
		logger.Error("Unable to unzip pak", "error", err)
		return err
	}

	return nil
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func DownloadTempFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	} else if resp.ContentLength <= 0 {
		return "", fmt.Errorf("empty response")
	}

	tempFile, err := os.CreateTemp("", "download-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func CreateTempQRCode(content string, size int) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)

	if err != nil {
		return "", err
	}

	qr.BackgroundColor = color.Black
	qr.ForegroundColor = color.White
	qr.DisableBorder = true

	tempFile, err := os.CreateTemp("", "qrcode-*")

	err = qr.Write(size, tempFile)

	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	return tempFile.Name(), err
}

func Unzip(src, dest string, pak models.Pak, isUpdate bool) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	extractAndWriteFile := func(f *zip.File) error {
		if isUpdate && ShouldIgnoreFile(f.Name, pak) {
			return nil
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return err
			}

			tempPath := path + ".tmp"
			tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}

			_, err = io.Copy(tempFile, rc)
			tempFile.Close()

			if err != nil {
				os.Remove(tempPath)
				return err
			}

			err = os.Rename(tempPath, path)
			if err != nil {
				os.Remove(tempPath)
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func ShouldIgnoreFile(filePath string, pak models.Pak) bool {
	for _, ignorePattern := range pak.UpdateIgnore {
		match, err := filepath.Match(ignorePattern, filePath)
		if err == nil && match {
			return true
		}

		parts := strings.Split(filePath, string(os.PathSeparator))
		for i := 0; i < len(parts); i++ {
			if i > 0 && strings.HasSuffix(parts[i-1], ".pak") {
				break
			}

			partialPath := strings.Join(parts[:i+1], string(os.PathSeparator))
			match, err := filepath.Match(ignorePattern, partialPath)
			if err == nil && match {
				return true
			}
		}
	}

	return false
}
