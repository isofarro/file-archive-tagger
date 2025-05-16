package fileops

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

// FileInfo represents metadata about a file
type FileInfo struct {
    Path       string
    Hash       string
    Size       int64
    ModifiedAt string
}

// CalculateFileHash computes SHA-256 hash of a file
func CalculateFileHash(path string) (string, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", fmt.Errorf("failed to calculate hash: %w", err)
    }

    return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetFileInfo returns file metadata including hash
func GetFileInfo(path string) (*FileInfo, error) {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return nil, fmt.Errorf("failed to get absolute path: %w", err)
    }

    stat, err := os.Stat(absPath)
    if err != nil {
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }

    hash, err := CalculateFileHash(absPath)
    if err != nil {
        return nil, err
    }

    return &FileInfo{
        Path:       absPath,
        Hash:       hash,
        Size:       stat.Size(),
        ModifiedAt: stat.ModTime().UTC().Format("2006-01-02 15:04:05"),
    }, nil
}

// CompareFiles checks if two files have the same content by comparing their hashes
func CompareFiles(path1, path2 string) (bool, error) {
    hash1, err := CalculateFileHash(path1)
    if err != nil {
        return false, err
    }

    hash2, err := CalculateFileHash(path2)
    if err != nil {
        return false, err
    }

    return hash1 == hash2, nil
}