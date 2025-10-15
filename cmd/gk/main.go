package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/neatflowcv/gk/internal/app/flow"
	"github.com/neatflowcv/gk/internal/pkg/filesystem/standard"
	"github.com/neatflowcv/gk/internal/pkg/kubernetes/printer"
)

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return info.Main.Version
}

type Arguments struct {
	Path string
}

func parseArguments() (*Arguments, error) {
	path := ""
	flag.StringVar(&path, "path", "", "순회할 루트 디렉토리 경로 (미지정 시 현재 작업 디렉토리)")
	flag.Parse()

	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("현재 작업 디렉토리 확인 실패: %w", err)
		}

		path = cwd
	}

	absRoot, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("경로(%v) 확인 실패: %w", path, err)
	}

	return &Arguments{
		Path: absRoot,
	}, nil
}

func main() {
	log.Println("version", version())

	args, err := parseArguments()
	if err != nil {
		log.Fatalf("인수 파싱 실패: %v", err)
	}

	ctx := context.Background()
	filesystem := standard.New()
	kubernetes := printer.NewKubernetes()
	service := flow.NewService(filesystem, kubernetes)

	err = service.ApplyPath(ctx, args.Path)
	if err != nil {
		log.Fatalf("적용 중 오류가 발생했습니다: %v", err)
	}
}
