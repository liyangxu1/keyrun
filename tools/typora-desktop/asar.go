package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type object = map[string]any

func patchArchive(data []byte, transform func([]byte, string, string) ([]byte, error)) ([]byte, error) {
	if len(data) < 16 || binary.LittleEndian.Uint32(data) != 4 {
		return nil, fmt.Errorf("无法识别 ASAR 格式")
	}
	headerSize := int(binary.LittleEndian.Uint32(data[4:]))
	jsonSize := int(binary.LittleEndian.Uint32(data[12:]))
	if headerSize < 8 || headerSize > len(data)-8 || jsonSize < 2 || jsonSize > headerSize-8 {
		return nil, fmt.Errorf("ASAR 索引长度无效")
	}
	var header object
	if err := json.Unmarshal(data[16:16+jsonSize], &header); err != nil {
		return nil, err
	}
	body := data[8+headerSize:]
	read := func(name string) (object, []byte, error) {
		node := header
		for _, segment := range strings.Split(name, "/") {
			files, _ := node["files"].(object)
			node, _ = files[segment].(object)
		}
		rawOffset, _ := node["offset"].(string)
		offset, err := strconv.Atoi(rawOffset)
		size, ok := node["size"].(float64)
		if err != nil || offset < 0 || offset > len(body) || !ok || size < 0 || size > float64(len(body)-offset) || float64(int(size)) != size {
			return nil, nil, fmt.Errorf("ASAR 缺少可读取的文件：%s", name)
		}
		return node, body[offset : offset+int(size)], nil
	}
	_, manifest, err := read("package.json")
	if err != nil {
		return nil, err
	}
	var pkg struct{ Main, Version string }
	if err := json.Unmarshal(manifest, &pkg); err != nil {
		return nil, err
	}
	entry := path.Clean(pkg.Main)
	if !strings.HasSuffix(entry, ".js") || path.IsAbs(entry) || entry == ".." || strings.HasPrefix(entry, "../") || strings.Contains(entry, "\\") {
		return nil, fmt.Errorf("不支持的 ASAR 入口：%s", entry)
	}
	node, original, err := read(entry)
	if err != nil {
		return nil, err
	}
	script, err := transform(original, entry, pkg.Version)
	if err != nil {
		return nil, err
	}
	hash := func(b []byte) string {
		sum := sha256.Sum256(b)
		return hex.EncodeToString(sum[:])
	}
	const blockSize = 4 << 20
	blocks := []string{}
	for start := 0; start < len(script); start += blockSize {
		blocks = append(blocks, hash(script[start:min(start+blockSize, len(script))]))
	}
	// 将新入口附在数据区末尾，其余条目的偏移及外置文件保持原样。
	node["offset"], node["size"] = strconv.Itoa(len(body)), len(script)
	node["integrity"] = object{"algorithm": "SHA256", "hash": hash(script), "blockSize": blockSize, "blocks": blocks}
	index, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	padded := (len(index) + 3) &^ 3
	out := make([]byte, 16+padded, 16+padded+len(body)+len(script))
	binary.LittleEndian.PutUint32(out, 4)
	binary.LittleEndian.PutUint32(out[4:], uint32(8+padded))
	binary.LittleEndian.PutUint32(out[8:], uint32(4+padded))
	binary.LittleEndian.PutUint32(out[12:], uint32(len(index)))
	copy(out[16:], index)
	return append(append(out, body...), script...), nil
}

func writeArchive(name string, data []byte, mode os.FileMode) error {
	return writeStaged(name, data, mode, false)
}

func writeStaged(name string, data []byte, mode os.FileMode, exclusive bool) error {
	f, err := os.CreateTemp(filepath.Dir(name), ".keyrun-typora-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(mode.Perm()); err == nil {
		_, err = f.Write(data)
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if exclusive {
		return os.Link(f.Name(), name)
	}
	return os.Rename(f.Name(), name)
}
