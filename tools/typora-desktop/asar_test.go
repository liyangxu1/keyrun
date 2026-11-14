package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"testing"
)

func TestPatchArchive(t *testing.T) {
	manifest := []byte(`{"main":"lib/main.js","version":"1.0.0"}`)
	source := []byte(`console.log("入口");`)
	header := object{"files": object{
		"package.json": object{"offset": "0", "size": len(manifest)},
		"lib": object{"files": object{
			"main.js":     object{"offset": strconv.Itoa(len(manifest)), "size": len(source)},
			"native.node": object{"unpacked": true, "size": 16},
		}},
	}}
	index, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}
	padded := (len(index) + 3) &^ 3
	input := make([]byte, 16+padded)
	binary.LittleEndian.PutUint32(input, 4)
	binary.LittleEndian.PutUint32(input[4:], uint32(8+padded))
	binary.LittleEndian.PutUint32(input[8:], uint32(4+padded))
	binary.LittleEndian.PutUint32(input[12:], uint32(len(index)))
	copy(input[16:], index)
	input = append(append(input, manifest...), source...)
	newEntry := []byte("changed")
	out, err := patchArchive(input, func(raw []byte, entry, version string) ([]byte, error) {
		if !bytes.Equal(raw, source) || entry != "lib/main.js" || version != "1.0.0" {
			t.Fatalf("入口解析错误：%q %s %s", raw, entry, version)
		}
		return newEntry, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(out, append(append(append([]byte{}, manifest...), source...), newEntry...)) {
		t.Fatal("原有数据或新入口内容不正确")
	}
	_, err = patchArchive(out, func(raw []byte, _, _ string) ([]byte, error) {
		if !bytes.Equal(raw, newEntry) {
			t.Fatal("更新后的索引未指向新入口")
		}
		return raw, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var result object
	if err := json.Unmarshal(out[16:16+binary.LittleEndian.Uint32(out[12:])], &result); err != nil {
		t.Fatal(err)
	}
	files := result["files"].(object)["lib"].(object)["files"].(object)
	if files["native.node"].(object)["unpacked"] != true {
		t.Fatal("外置文件标志丢失")
	}
	if _, err := patchArchive(input[:16], nil); err == nil {
		t.Fatal("截断的 ASAR 索引未被拒绝")
	}
}
