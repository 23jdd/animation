// Package fontasset provides fonts compiled into the application binary.
package fontasset

import _ "embed"

// JetBrains Mono Regular is sourced from the official JetBrainsMono repository
// at revision 19371302b95d218af43299bce79ddbddd0bc364d.
//
//go:embed assets/JetBrainsMono-Regular.ttf
var jetBrainsMonoRegular []byte

//go:embed assets/JetBrainsMono-OFL.txt
var jetBrainsMonoLicense string

// Source Han Sans SC Regular is sourced from Adobe's official Source Han Sans
// 2.005R release and is the proportional default for animated scene text.
//
//go:embed assets/SourceHanSansSC-Regular.otf
var sourceHanSansSCRegular []byte

//go:embed assets/SourceHan-OFL.txt
var sourceHanSansLicense string

// JetBrainsMonoRegular returns the embedded JetBrains Mono Regular font data.
// Callers must treat the returned bytes as read-only.
func JetBrainsMonoRegular() []byte { return jetBrainsMonoRegular }

// JetBrainsMonoLicense returns the embedded SIL Open Font License 1.1 text.
func JetBrainsMonoLicense() string { return jetBrainsMonoLicense }

// SourceHanSansSCRegular returns the embedded Source Han Sans SC Regular font.
// Callers must treat the returned bytes as read-only.
func SourceHanSansSCRegular() []byte { return sourceHanSansSCRegular }

// SourceHanSansLicense returns the embedded SIL Open Font License 1.1 text.
func SourceHanSansLicense() string { return sourceHanSansLicense }
