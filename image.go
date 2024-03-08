package main

import (
	"fmt"
	"strings"
)

type Image struct {
	url       string
	registry  string
	namespace string
	repo      string
	tag       string
}

func ParseImage(url string) (*Image, error) {
	// registry/namespace/repo:tag or namespace/repo:tag or repo:tag
	slice := strings.SplitN(url, "/", 3)

	var tag, repo string

	// 取最后一个
	repoAndTag := slice[len(slice)-1]
	// 以:分割 Repo 和 Tag
	s := strings.Split(repoAndTag, ":")
	// 如果长度大于2，说明不是一个合法的镜像地址
	if len(s) > 2 {
		return nil, fmt.Errorf("invalid image url: %v", url)
	}

	// 如果长度等于2，说明有tag
	if len(s) == 2 {
		repo = s[0]
		tag = s[1]
	}

	// 如果长度等于1，说明没有tag
	if len(s) == 1 {
		repo = s[0]
		tag = ""
	}

	// 如果长度等于3，说明格式为registry/namespace/repo:tag
	if len(slice) == 3 {
		return &Image{
			url:       url,
			registry:  slice[0],
			namespace: slice[1],
			repo:      repo,
			tag:       tag,
		}, nil
	}

	// 如果长度等于2，说明格式为 namespace/repo:tag 或者 domain/repo:tag
	if len(slice) == 2 {
		// if first string is a domain
		if strings.Contains(slice[0], ".") {
			return &Image{
				url:       url,
				registry:  slice[0],
				namespace: "",
				repo:      repo,
				tag:       tag,
			}, nil
		}

		return &Image{
			url:       url,
			registry:  "docker.io",
			namespace: slice[0],
			repo:      repo,
			tag:       tag,
		}, nil
	}

	// 如果长度等于1，说明格式为 repo:tag
	return &Image{
		url:       url,
		registry:  "docker.io",
		namespace: "library",
		repo:      repo,
		tag:       tag,
	}, nil
}

// GetURL returns the whole url
func (r *Image) GetURL() string {
	url := r.GetURLWithoutTag()
	if r.tag != "" {
		url = url + ":" + r.tag
	}
	return url
}

// GetOriginURL returns the whole url
func (r *Image) GetOriginURL() string {
	return r.url
}

// GetRegistry returns the registry in a url
func (r *Image) GetRegistry() string {
	return r.registry
}

// GetNamespace returns the namespace in a url
func (r *Image) GetNamespace() string {
	return r.namespace
}

// GetRepo returns the repository in a url
func (r *Image) GetRepo() string {
	return r.repo
}

// GetTag returns the tag in a url
func (r *Image) GetTag() string {
	return r.tag
}

// GetRepoWithNamespace returns namespace/repository in a url
func (r *Image) GetRepoWithNamespace() string {
	if r.namespace == "" {
		return r.repo
	}
	return r.namespace + "/" + r.repo
}

// GetRepoWithTag returns repository:tag in a url
func (r *Image) GetRepoWithTag() string {
	if r.tag == "" {
		return r.repo
	}
	return r.repo + ":" + r.tag
}

// GetURLWithoutTag returns registry/namespace/repository in a url
func (r *Image) GetURLWithoutTag() string {
	if r.namespace == "" {
		return r.registry + "/" + r.repo
	}
	return r.registry + "/" + r.namespace + "/" + r.repo
}
