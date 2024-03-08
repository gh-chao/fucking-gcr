package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"
)

//
// 1. LoadJsonBlocksFromStdin 从stdin中读取yaml, 返回一个json数组
// 2. ProcessJsonBlock 处理json数组，替换其中的镜像地址，返回替换后的json数组
// 3. 将替换后的json数组转换为yaml，输出到stdout
//

var (
	mirror     string
	whitelist  string
	scriptName string
)

func init() {
	flag.StringVar(&mirror, "mirror", "registry.baidubce.com/fucking-gcr", "指定镜像地址")
	flag.StringVar(&whitelist, "whitelist", "gcr.io,docker.io", "指定白名单")
	flag.StringVar(&scriptName, "script-name", "run-skopeo-copy.sh", "镜像同步脚本名称")
	flag.Parse()
}

func main() {
	jsonBlocks := LoadJsonBlocksFromStdin()

	results := make(map[string]string)

	for _, jsonBlock := range jsonBlocks {
		changed, m, err := ProcessJsonBlock(mirror, strings.Split(whitelist, ","), jsonBlock)
		if err != nil {
			panic(err)
		}

		for k, v := range m {
			results[k] = v
		}

		DumpJsonBlockToStdout(changed)
	}

	WriteImageCopyScript(scriptName, results)
}

func LoadJsonBlocksFromStream(r io.Reader) []string {
	localBuilder := resource.NewLocalBuilder()

	result := localBuilder.Unstructured().
		ContinueOnError().
		Stream(r, "").
		Flatten().
		Do()

	infos, err := result.Infos()
	if err != nil {
		panic(err)
	}

	blocks := make([]string, 0, len(infos))

	for _, info := range infos {
		bytes, err := info.Object.(*unstructured.Unstructured).MarshalJSON()
		if err != nil {
			panic(err)
		}
		blocks = append(blocks, string(bytes))
	}

	return blocks
}

func LoadJsonBlocksFromStdin() []string {
	return LoadJsonBlocksFromStream(os.Stdin)
}

func DumpJsonBlockToStdout(block string) {
	// 转换为yaml
	yamlBlock, err := yaml.JSONToYAML([]byte(block))
	if err != nil {
		panic(err)
	}
	fmt.Println("---")
	fmt.Println(string(yamlBlock))
}

func WriteImageCopyScript(name string, results map[string]string) {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for origin, dist := range results {
		fmt.Fprintf(f, "skopeo copy --multi-arch=all docker://%s docker://%s\n", origin, dist)
	}
}

func ProcessJsonBlock(mirror string, whitelist []string, in string) (string, map[string]string, error) {
	out := in

	results := make(map[string]string)
	images := FindImages(in)

	for _, s := range whitelist {
		images = append(images, FindImageWithRegistry(s, in)...)
	}

	for _, origin := range images {
		dest, err := RenameImage(mirror, whitelist, origin)
		if err != nil {
			return "", results, err
		}

		if len(dest) == 0 {
			continue
		}

		out = strings.ReplaceAll(out, fmt.Sprintf(`"%s"`, origin), fmt.Sprintf(`"%s"`, dest))

		results[origin] = dest
	}

	return out, results, nil
}

// FindImages 在字符串中查找key=image的镜像地址
func FindImages(s string) []string {
	var ImageRegexp = regexp.MustCompile(fmt.Sprintf(`"image":"%s"`, ReferenceRegexp))
	images := ImageRegexp.FindAllString(s, -1)
	for k, v := range images {
		images[k] = strings.TrimSuffix(strings.TrimPrefix(v, `"image":"`), `"`)
	}

	return images
}

func FindImageWithRegistry(registry, s string) []string {
	var ImageRegexp = regexp.MustCompile(fmt.Sprintf(`"%s/%s"`, registry, ReferenceRegexp))
	images := ImageRegexp.FindAllString(s, -1)
	for k, v := range images {
		images[k] = strings.Trim(v, `"`)
	}

	return images
}

// RenameImage 重命名镜像地址
func RenameImage(mirror string, whitelist []string, imageURL string) (string, error) {
	// 将k8s.gcr.io/替换为gcr.io/google_containers/
	if strings.HasPrefix(imageURL, "k8s.gcr.io/") {
		imageURL = strings.Replace(imageURL, "k8s.gcr.io/", "gcr.io/google_containers/", 1)
	}

	var sha256 string
	// 如果镜像地址中包含@sha256，去掉
	if strings.Contains(imageURL, "@sha256:") {
		s := strings.Split(imageURL, "@sha256:")
		imageURL = s[0]
		sha256 = s[1]
	}

	image, err := ParseImage(imageURL)
	if err != nil {
		return "", err
	}

	// 如果镜像地址中包含sha256，取sha256前16位当做tag
	if len(sha256) > 0 {
		image.tag = sha256[:16]
	}

	if !CheckWhitelist(whitelist, image.GetURL()) {
		return "", nil
	}

	// 替换镜像地址
	dest := fmt.Sprintf("%s/%s:%s",
		strings.TrimRight(mirror, "/"),
		image.GetRegistry()+"/"+image.GetRepoWithNamespace(),
		image.tag,
	)

	return dest, nil
}

func CheckWhitelist(whitelist []string, imageURL string) bool {
	for _, v := range whitelist {
		if strings.HasPrefix(imageURL, v) {
			return true
		}
	}
	return false
}
