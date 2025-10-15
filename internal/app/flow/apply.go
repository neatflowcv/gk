package flow

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ApplyPath는 루트 경로의 1레벨 하위 디렉토리를 네임스페이스로 간주하고,
// 각 디렉토리 내의 모든 YAML 파일을 kubectl apply 한다.
// 반환값은 실패한 파일 수이다.
func ApplyPath(ctx context.Context, rootPath string) error {
	namespaces, err := listNamespaces(rootPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "디렉토리 나열 실패: %v\n", err)

		return err
	}

	totalFailures := 0
	successCount := 0

	for _, namespaceDir := range namespaces {
		namespaceName := filepath.Base(namespaceDir)

		err := ensureNamespace(ctx, namespaceName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "네임스페이스 보장 실패 [%s]: %v\n", namespaceName, err)
			totalFailures++

			// 네임스페이스 보장에 실패했으면 해당 네임스페이스는 건너뛴다
			continue
		}

		yamlFiles, err := collectYamlFiles(namespaceDir)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "YAML 수집 실패 [%s]: %v\n", namespaceName, err)
			totalFailures++

			continue
		}

		for _, yamlFile := range yamlFiles {
			err := kubectlApply(ctx, namespaceName, yamlFile)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "apply 실패 [%s] %s: %v\n", namespaceName, yamlFile, err)
				totalFailures++
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "apply 성공 [%s] %s\n", namespaceName, yamlFile)
				successCount++
			}
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "요약: 네임스페이스 %d개, 적용 성공 %d, 실패 %d\n", len(namespaces), successCount, totalFailures)

	return fmt.Errorf("%w: %d개", ErrApplyFailures, totalFailures)
}

var ErrNoNamespaceDirs = errors.New("네임스페이스 디렉토리가 없습니다")
var ErrApplyFailures = errors.New("적용 실패")

func listNamespaces(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("루트 디렉토리 읽기 실패: %w", err)
	}

	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(root, entry.Name()))
		}
	}

	if len(dirs) == 0 {
		return nil, ErrNoNamespaceDirs
	}

	return dirs, nil
}

func collectYamlFiles(dir string) ([]string, error) {
	var files []string

	walkErr := filepath.WalkDir(dir, func(path string, dirEntry os.DirEntry, pathErr error) error {
		if pathErr != nil {
			return pathErr
		}

		if dirEntry.IsDir() {
			return nil
		}

		lowerName := strings.ToLower(dirEntry.Name())
		if strings.HasSuffix(lowerName, ".yaml") || strings.HasSuffix(lowerName, ".yml") {
			files = append(files, path)
		}

		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("YAML 파일 순회 실패(%s): %w", dir, walkErr)
	}

	return files, nil
}

func ensureNamespace(ctx context.Context, namespaceName string) error {
	// kubectl get ns <ns>
	getErr := runCmd(ctx, "kubectl", []string{"get", "ns", namespaceName})
	if getErr == nil {
		return nil
	}

	// create
	return runCmd(ctx, "kubectl", []string{"create", "ns", namespaceName})
}

func kubectlApply(ctx context.Context, namespaceName, file string) error {
	return runCmd(ctx, "kubectl", []string{"apply", "-n", namespaceName, "-f", file})
}

func runCmd(ctx context.Context, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("명령 시작 실패 (%s %s): %w", name, strings.Join(args, " "), err)
	}

	go pipeLines(stdout, os.Stdout)
	go pipeLines(stderr, os.Stderr)

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("명령 실행 실패 (%s %s): %w", name, strings.Join(args, " "), err)
	}

	return nil
}

func pipeLines(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		_, _ = fmt.Fprintln(w, scanner.Text())
	}
}
