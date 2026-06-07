package mmkv

import (
	"fmt"
	"os"

	"git.um-react.app/um/cli/algo/common"
	"git.um-react.app/um/cli/internal/utils"
	go_mmkv "github.com/unlock-music/go-mmkv"
	"go.uber.org/zap"
)

func Merge(keys ...common.QMCKeys) common.QMCKeys {
	result := make(common.QMCKeys)
	for _, k := range keys {
		for key, value := range k {
			result[utils.NormalizeUnicode(key)] = utils.NormalizeUnicode(value)
		}
	}
	return result
}

func LoadFromPath(path string, key string, logger *zap.Logger) (result common.QMCKeys, err error) {
	mmkv_path := path
	mmkv_crc := path + ".crc"

	mr, err := os.Open(mmkv_path)
	if err != nil {
		logger.Error("LoadMMKV: Could not open mmkv file", zap.Error(err))
		return nil, fmt.Errorf("LoadMMKV: open error: %w", err)
	}
	defer mr.Close()

	cr, err := os.Open(mmkv_crc)
	if err != nil {
		// The .crc file carries the MMKV encryption metadata; without it the store
		// can only be read as unencrypted. Make it explicit when a supplied key is
		// being ignored as a result.
		if key != "" {
			logger.Warn("LoadMMKV: crc file missing, ignoring provided key and assuming no encryption", zap.Error(err))
		} else {
			logger.Warn("LoadMMKV: crc file missing, assuming no encryption", zap.Error(err))
		}
		key = ""
	} else {
		defer cr.Close()
	}

	var password []byte = nil
	if key != "" {
		if len(key) != 16 {
			logger.Warn("LoadMMKV: encryption key is not 16 bytes; it will be zero-padded or truncated, which may cause decryption to fail",
				zap.Int("keyLength", len(key)))
		}
		password = make([]byte, 16)
		copy(password, []byte(key))
	}
	mmkv, err := go_mmkv.NewMMKVReader(mr, password, cr)
	if err != nil {
		logger.Error("LoadMMKV: failed to create reader", zap.Error(err))
		return nil, fmt.Errorf("LoadMMKV: NewMMKVReader error: %w", err)
	}

	result = make(common.QMCKeys)
	// maxEntries is a backstop against a malformed store that never reaches EOF
	// while ReadKey/ReadStringValue keep returning without error. Far above any
	// realistic QQ Music key store.
	const maxEntries = 1 << 20
	for count := 0; !mmkv.IsEOF(); count++ {
		if count >= maxEntries {
			logger.Warn("LoadMMKV: aborting after too many entries; file may be malformed", zap.Int("limit", maxEntries))
			break
		}
		key, err := mmkv.ReadKey()
		if err != nil {
			logger.Error("LoadMMKV: read key error", zap.Error(err))
			return nil, fmt.Errorf("LoadMMKV: read key error: %w", err)
		}
		value, err := mmkv.ReadStringValue()
		if err != nil {
			logger.Error("LoadMMKV: read value error", zap.Error(err))
			return nil, fmt.Errorf("LoadMMKV: read value error: %w", err)
		}
		logger.Debug("LoadMMKV: read", zap.String("key", key), zap.String("value", value))
		result[utils.NormalizeUnicode(key)] = utils.NormalizeUnicode(value)
	}

	return result, nil
}
