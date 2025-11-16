package testutil

import "os/exec"

func CreateTestMP4(path string) error {
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "color=c=black:s=480x360:d=1", "-f", "lavfi", "-i", "anullsrc", "-shortest", "-c:v", "libx264", "-c:a", "aac", path, "-y")
	return cmd.Run()
}
