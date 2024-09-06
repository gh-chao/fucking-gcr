# fucking-gcr

解决下载 gcr.io 等国外镜像源超时的问题, 使用工具自动替换 k8s yaml 里的镜像地址

1. 从 stdin 读取 k8s yaml
2. 自动替换镜像地址
3. 输出替换后的yaml 到stdout
4. 将镜像同步命令写入到 copy-image.sh


### 依赖

* 镜像同步工具 [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md)

使用 brew 安装
```bash
brew install crane
```

### 安装

使用 make 安装
```bash
make
make install
```

手动安装
```bash
./gow build -o bin/fucking-gcr .
cp bin/fucking-gcr /usr/local/bin
```

### 参数
```bash
-mirror string
    指定镜像地址 (default "registry.baidubce.com/fucking-gcr")
-script-name string
    镜像同步脚本名称 (default "copy-image.sh")
-whitelist string
    指定白名单 (default "gcr.io,docker.io")
```

### 使用

以 wordpress 为例

```bash
# 下载argocd安装资源
curl -s https://kubernetes.io/examples/application/wordpress/wordpress-deployment.yaml > wordpress-deployment.yaml

# 替换 install.yaml 里的地址
cat wordpress-deployment.yaml | fucking-gcr --mirror=registry.baidubce.com/fucking-gcr --whitelist=gcr.io,docker.io > wordpress-deployment-patched.yaml

# 比较差异
brew install homeport/tap/dyff
dyff bw wordpress-deployment.yaml wordpress-deployment-patched.yaml

# 将镜像同步到国内源（这一步需要使用代理，你懂的）
https_proxy=127.0.0.1:1234 ./copy-image.sh

# 安装
kubectl apply -f wordpress-deployment-patched.yaml
```

单行命令

```bash
curl -s https://kubernetes.io/examples/application/wordpress/wordpress-deployment.yaml \
| fucking-gcr --mirror=registry.baidubce.com/fucking-gcr --whitelist=gcr.io,docker.io \
| kubectl apply -f -
```

### 在helm中使用

```bash
# 添加 --post-renderer 和  --post-renderer-args 参数即可
helm template wordpress bitnami/wordpress \
    --post-renderer fucking-gcr \
    --post-renderer-args "--whitelist=gcr.io,docker.io" \
    --post-renderer-args "--mirror=registry.baidubce.com/fucking-gcr"
```
