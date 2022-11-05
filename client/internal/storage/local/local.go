package local

import (
	"fmt"
	"os"
)

var (
	dirPath   = "thumbnail_jpg"
	filePaths = "/file%v.jpg"
)

func SaveImg(imgs [][]byte) error {
	if imgs == nil {
		return fmt.Errorf("in SaveImg imgs == nil")
	}
	err := os.Mkdir(dirPath, 0777)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("in SaveImg could not Mkdir %s: %w", dirPath, err)
	}
	for index, img := range imgs {
		if img == nil {
			if err == nil {
				err = fmt.Errorf("in SaveImg [%d] image == nil", index)
			} else {
				err = fmt.Errorf("and [%d] image == nil", index)
			}
		}

		path := fmt.Sprintf(dirPath+filePaths, index)
		out, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("in SaveImg could not Create %s: %w", path, err)
		}

		_, err = out.Write(img)
		if err != nil {
			return fmt.Errorf("in SaveImg can`t write in %s: %w", path, err)
		}

		err = out.Close()
		if err != nil {
			return fmt.Errorf("in SaveImg can`t close %s: %w", path, err)
		}
	}

	return err
}
