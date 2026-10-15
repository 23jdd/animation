package script

import "os"

func Load(path string) (string, error) { b, err := os.ReadFile(path); return string(b), err }
