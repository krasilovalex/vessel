package vessel

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
)

type Builder struct {
	tag        string
	directives []string          // instruction RUN, WORKDIR
	copies     map[string]string // Mapping file for COPY (src -> dest)
	copyDirs   map[string]string // directory (source_dir -> dest_dir)
	err        error
}

// initalizes the builder
func NewBuilder(tag string) *Builder {
	if tag == "" {
		return &Builder{err: fmt.Errorf("vessel: image tag cannot be empty")}
	}

	return &Builder{
		tag:      tag,
		copies:   make(map[string]string),
		copyDirs: make(map[string]string),
	}
}

// From sets the base image
func (b *Builder) From(image string) *Builder {
	if b.err != nil {
		return b
	}

	b.directives = append(b.directives, fmt.Sprintf("FROM %s", image))
	return b
}

// Run adds a command to execute
func (b *Builder) Run(cmd string) *Builder {
	if b.err != nil {
		return b
	}

	b.directives = append(b.directives, fmt.Sprintf("RUN %s", cmd))
	return b
}

// WORKDIR changes the working directory
func (b *Builder) Workdir(dir string) *Builder {
	if b.err != nil {
		return b
	}
	b.directives = append(b.directives, fmt.Sprintf("WORKDIR %s", dir))
	return b
}

// Copy adds the file to the build context
func (b *Builder) Copy(src, dest string) *Builder {
	if b.err != nil {
		return b
	}
	b.copies[src] = dest
	b.directives = append(b.directives, fmt.Sprintf("COPY %s %s", src, dest))
	return b
}

// CopyFrom copies files from the specifed previous bulid step
func (b *Builder) CopyFrom(stage, src, dest string) *Builder {
	if b.err != nil {
		return b
	}

	b.directives = append(b.directives, fmt.Sprintf("COPY --from=%s %s %s", stage, src, dest))
	return b
}

// CopyDir adds the entire directory to the build context recursively
func (b *Builder) CopyDir(srcDir, destDir string) *Builder {
	if b.err != nil {
		return b
	}

	b.copyDirs[srcDir] = destDir
	b.directives = append(b.directives, fmt.Sprintf("COPY %s %s", srcDir, destDir))
	return b
}

// Build creates an in-memory tar context and sends it to the Docker API
func (b *Builder) Build(ctx context.Context) error {
	if b.err != nil {
		return fmt.Errorf("fluent error: %w", b.err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to init docker client: %w", err)
	}

	dockerfile := ""
	for _, d := range b.directives {
		dockerfile += d + "\n"
	}

	pterm.Info.Printfln("Bulding image '%s' with dynamic context...", b.tag)
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name: "Dockerfile",
		Mode: 0600,
		Size: int64(len(dockerfile)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		return err
	}

	for src := range b.copies {
		content, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read file %s for context: %w", src, err)
		}
		fileHdr := &tar.Header{
			Name: src,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(fileHdr); err != nil {
			return err
		}

		if _, err := tw.Write(content); err != nil {
			return err
		}
	}

	for srcDir := range b.copyDirs {
		err := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(srcDir, path)
			if err != nil {
				return err
			}

			archivePath := fmt.Sprintf("%s/%s", srcDir, filepath.ToSlash(relPath))

			content, err := os.ReadFile(path)

			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			fileHdr := &tar.Header{
				Name: archivePath,
				Mode: 0644,
				Size: int64(len(content)),
			}

			if err := tw.WriteHeader(fileHdr); err != nil {
				return err
			}

			if _, err := tw.Write(content); err != nil {
				return err
			}

			return err
		})

		if err != nil {
			return fmt.Errorf("failed to process directory %s: %w", srcDir, err)
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	resp, err := cli.ImageBuild(ctx, buf, types.ImageBuildOptions{
		Tags:       []string{b.tag},
		Dockerfile: "Dockerfile",
		Remove:     true,
	})

	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			break // Поток закончился, сборка завершена
		} else if err != nil {
			return fmt.Errorf("failed to parse docker output: %w", err)
		}

		// Если Docker прислал ошибку - прерываем процесс и возвращаем ее
		if errMsg, ok := msg["error"].(string); ok {
			return fmt.Errorf("build failed: %s", errMsg)
		}

		// Красиво выводим логи (stream - обычные логи, status - прогресс скачивания)

		if stream, ok := msg["stream"].(string); ok {
			pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.FgGray)).Print(stream)
		}
	}
	pterm.Success.Printfln("Image %s successfully built", b.tag)
	return nil
}

// cmd
func (b *Builder) Cmd(cmd string) *Builder {
	if b.err != nil {
		return b
	}

	b.directives = append(b.directives, fmt.Sprintf(`CMD ["sh", "-c", "%s"]`, cmd))
	return b
}
