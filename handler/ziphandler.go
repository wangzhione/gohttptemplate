package handler

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wangzhione/sbp/util/filedir"
)

func REQUEST[REQ any](logic func(context.Context, REQ, http.ResponseWriter)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := r.Context()

		var request REQ
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			// 如果恶意尝试过多, 会关闭
			slog.InfoContext(c, "REQUEST json.NewDecoder(r.Body).Decode(&request) error",
				slog.Any("error", err),
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
			)
			ResponseWriterMethodError(w, http.StatusBadRequest)
			return
		}

		logic(c, request, w)
	}
}

// ResponseWriterZip 打包并将文件流式返回为 ZIP 下载
func ResponseWriterZip(ctx context.Context, w http.ResponseWriter, zipname string, files []string) error {
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			slog.ErrorContext(ctx, "Stat zip file failed", "error", err, "file", file)
			return err
		}
		if info.IsDir() {
			err := fmt.Errorf("zip file path is directory: %s", file)
			slog.ErrorContext(ctx, "Zip file path is directory", "error", err, "file", file)
			return err
		}
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", "attachment; filename=\""+zipname+"\"")
	w.Header().Set("Content-Type", "application/zip")

	zipWriter := zip.NewWriter(w)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			slog.ErrorContext(ctx, "Failed to close zipWriter", "error", err, "zipname", zipname)
		}
	}()

	for _, file := range files {
		err := filedir.AddFileToZip(ctx, zipWriter, file, filepath.Base(file))
		if err != nil {
			return err
		}
	}

	return nil
}

func ResponseWriterZipDir(ctx context.Context, w http.ResponseWriter, zipname string, dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		slog.ErrorContext(ctx, "Stat zip dir failed", "error", err, "dir", dir)
		return err
	}
	if !info.IsDir() {
		err := fmt.Errorf("zip dir path is not directory: %s", dir)
		slog.ErrorContext(ctx, "Zip dir path is not directory", "error", err, "dir", dir)
		return err
	}

	// 设置响应头，提前发送 attachment 信息
	w.Header().Set("Content-Disposition", "attachment; filename="+zipname)
	w.Header().Set("Content-Type", "application/zip")

	// 创建 ZIP writer，直接写入 HTTP 响应流
	zipWriter := zip.NewWriter(w)
	defer func() {
		// 尝试关闭 zipWriter，刷新缓存
		if err := zipWriter.Close(); err != nil {
			// 如果 zipWriter 关闭失败
			slog.ErrorContext(ctx, "Failed to finalize zipWriter", "error", err, "zipname", zipname)
			return
		}
	}()

	// 遍历目录
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.ErrorContext(ctx, "Walk error", "error", err, "path", path)
			return err
		}

		// 构造 zip 文件中的相对路径（保留目录结构）
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			slog.ErrorContext(ctx, "Path Rel error", "error", err, "path", path)
			return err
		}

		// 忽略根路径（.）
		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			// 是目录则创建空目录条目
			_, err := zipWriter.Create(relPath + "/")
			if err != nil {
				slog.ErrorContext(ctx, "Create zip folder entry failed", "error", err, "folder", relPath)
				return err
			}
			return nil
		}

		// 文件处理
		return filedir.AddFileToZip(ctx, zipWriter, path, relPath)
	})
	// 如果 Walk 过程中出错
	if err != nil {
		slog.ErrorContext(ctx, "Failed to walk and zip directory", "error", err, "dir", dir, "zipname", zipname)
		// 注意：不能再使用 http.Error，因为响应头已发，部分内容已写入
		return err
	}

	return nil
}
