package images

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/waynezhang/foto/internal/config"
	"github.com/waynezhang/foto/internal/files"
	"github.com/waynezhang/foto/internal/log"
)

type Section struct {
  Title string
  Text string
  Slug string
  Folder string
  ImageSets []ImageSet
}

type ImageSet struct {
  FileName string
  ThumbnailSize ImageSize
  OriginalSize ImageSize
}

func ExtractPhotos(cfg config.Config, outputFolder *string) []Section {
  if cfg["section"] == nil {
    return []Section{}
  }
  sections := []Section{}
  for _, val := range cfg["section"].([]interface{}) {
    s := extractSection(val.(map[string]interface{}), cfg.GetExtractOption(), outputFolder)
    sections = append(sections, s)
  }

  return sections
}

func extractSection(info map[string]interface{}, option config.ExtractOption, outputPath *string) (Section) {
  title := info["title"].(string)
  text := info["text"].(string)
  slug := info["slug"].(string)
  folder := info["folder"].(string)
  log.Debug("Extacting section [%s][/%s] %s", title, slug, folder)

  imageSet := []ImageSet{}
	err := filepath.WalkDir(folder, func(path string, info os.DirEntry, err error) error {
    if err != nil {
      return err
    }
    if info.IsDir() || !IsPhotoSupported(path) {
      return nil
    }

    log.Debug("Processing image %s", path)
    imgSet, err := extractImage(path, option, slug, outputPath)
    if err != nil {
      return err
    }
    imageSet = append(imageSet, *imgSet)

		return nil
  })
  if err != nil {
    log.Fatal("Failed to get photos from %s (%v)", folder, err)
  }
  
  sort.SliceStable(imageSet, func(i, j int) bool {
    return imageSet[i].FileName > imageSet[j].FileName
  })

  return Section {
    title,
    text,
    slug,
    folder,
    imageSet,
  }
}

func extractImage(path string, option config.ExtractOption, slug string, outputPath *string) (*ImageSet, error) {
  imageSize, err := GetPhotoSize(path)
  if err != nil {
    return nil, err
  }

  ratio := float32(imageSize.Height) / float32(imageSize.Width)

  thumbnailWidth := option.ThumbnailWidth
  thumbnailHeight := int(float32(thumbnailWidth) * ratio)

  originalWidth := option.OriginalWidth
  originalHeight := int(float32(originalWidth) * ratio)

  if outputPath != nil {
    originalPath := files.OutputPhotoOriginalFilePath(*outputPath, slug, path)
    if err := ResizeImage(path, originalWidth, originalPath); err != nil {
      return nil, err
    }

    thumbnailPath := files.OutputPhotoThumbnailFilePath(*outputPath, slug, path)
    if err := ResizeImage(path, thumbnailWidth, thumbnailPath); err != nil {
      return nil, err
    }
  }

  return &ImageSet {
    FileName: filepath.Base(path),
    ThumbnailSize: ImageSize {
      thumbnailWidth,
      thumbnailHeight,
    },
    OriginalSize: ImageSize {
      originalWidth,
      originalHeight,
    },
  }, nil
}