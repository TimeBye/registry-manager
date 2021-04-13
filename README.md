# registry-manager

此程序调用 Registry API 以及 [skopeo](https://github.com/containers/skopeo#skopeo-) ，按照镜像 `tag` 中是否包含定义的关键字或匹配的正则表达式进行删除或同步操作。删除操作只进行软删除，不回收镜像实际所占物理存储。

⚠️ **警告** ⚠️ 如果`tag A`和`tag B`都指向同一个`image`，那么当你在删除`tag A`时，`tag B`也将被删除。

### 使用registry-manager进行软删除

- 编写配置文件`config.yml`
```yaml
# 所有镜像仓库地址(必填属性)
registries:
  # 镜像仓库名称，自定义即可
  harbor:
    # 镜像仓库地址
    registry: https://harbor.io
    # 镜像仓库管理员用户名
    username: admin
    # 镜像仓库管理员密码
    password: harbor12345
    # 是否跳过证书检查(默认：false)
    insecure: true
    # 需要删除/同步的项目名，不写则获取所有
    repositories:
      # 若指定tag，则只删除/同步该镜像
      - library/redis:5.0.12
      # 未指定tag，则删除/同步所有tag
      - library/mysql
  aliyun:
    registry: https://registry.aliyun.com
    username: admin
    password: harbor12345
    insecure: false

# 删除策略(可选属性，只进行镜像同步则可以不写删除策略)
delete-policy:
  # 需要删除的仓库
  registries:
    - harbor
    - aliyun
  # 若执行中断，指定起始仓库序号，默认 0
  start: 0
  # 仅模拟运行，不真实删除，默认启用
  dry-run: true
  # 保留会被删除的镜像个数，默认 10
  mix-count: 10
  # 是否删除不符合语义化版本的 tag
  sem-ver: false
  # 镜像tag删除策略
  tags:
    # 删除策略
    include:
      # 按关键字进行删除
      keys: develop,release,hotfix,feature,2018,2019
      # 按正则表达式删除
      regex:
    # 排除策略，删除策略与排除策略都匹配，以排除策略为准
    exclude:
      # 按关键字进行排除
      keys:
      # 按正则表达式排除
      regex: ^latest$|^master$|^[Vv]?(\d+(\.\d+){1,2})(-alpha\.\d+)?$

# 同步策略(可选属性，只进行镜像删除则可以不写同步策略)
sync-policy:
  # 源仓库
  from: harbor
  # 目标仓库
  to: aliyun
  # 若执行中断，指定起始仓库序号，默认 0
  start: 0
  # 仅模拟运行，不真实删除，默认启用
  dry-run: true
  # 替换规则
  replace:
    - old: library
      new: prod
  # 需要同步的tag筛选规则
  filters:
    - '^[Vv]?(\d+(\.\d+){1,2})(-alpha\.\d+)?$'
```

- 使用 docker 命令运行

```bash
# 删除镜像 tag
docker run -v $PWD/config.yml:/config.yml \
    setzero/registry-manager registry-manager delete -c /config.yml

# 同步镜像
docker run -v $PWD/config.yml:/config.yml \
    setzero/registry-manager registry-manager sync -c /config.yml

# 列出镜像库中所有镜像列表
docker run -v $PWD/config.yml:/config.yml \
    setzero/registry-manager registry-manager list -c /config.yml
```

### 存储回收

#### Harbor v1.7.0及以上版本

Harbor从v1.7.0版本开始支持不停机进行[在线存储回收](https://github.com/goharbor/harbor/blob/v1.7.0/docs/user_guide.md#online-garbage-collection)。在调用本程序进行软删除后，系统管理员可以通过单击“管理”下“配置”部分的“垃圾回收”选项卡来配置或触发存储回收。

👋 **注意** 👋 在执行存储回收时，Harbor将进入只读模式，并且禁止对 docker registry 进行任何修改。换而言之就是此时只能拉镜像不能推镜像。

#### Harbor 1.7.0以前版本

Harbor v1.7.0 以前版本进行存储回收时需要手动切断外部访问以达到`禁止对 docker registry 进行任何修改`的目的。回收镜像所占存储[参考文档](https://github.com/docker/docker.github.io/blob/v18.09-release/registry/garbage-collection.md#about-garbage-collection)。

- 切断外部访问入口
- 进入到`registry`容器中执行存储回收命令

  ```console
  # 测试回收，不会真回收，可在日志中看到要回收的镜像
  $ registry garbage-collect --dry-run /etc/registry/config.yml
  # 执行回收，没有后悔药
  $ registry garbage-collect /etc/registry/config.yml
  ```

#### 不理想的地方

不论是哪个版本的Harbor进行存储回收都是使用`docker registry`官方的命令进行回收，但回收空间太少，很多manifests仍没删除。那就只有扫描镜像仓库存储文件，通过`docker registry api`删除无用的manifests。这里可参考使用`mortensrasmussen`的[docker-registry-manifest-cleanup](https://hub.docker.com/r/mortensrasmussen/docker-registry-manifest-cleanup/)项目。

- 使用docker-registry-manifest-cleanup当前最新版本进行存储回收
  ```console
  # 执行以下脚本尝试通过api模拟删除manifests
  $ docker run -it \
      -v /home/someuser/registry:/registry \
      -e REGISTRY_URL=https://registry.example.com \
      -e DRY_RUN="true" \
      -e SELF_SIGNED_CERT="true" \
      -e REGISTRY_AUTH="myuser:sickpassword" \
      mortensrasmussen/docker-registry-manifest-cleanup:1.1.1
  # 如上一步没有报错，执行以下脚本，真正删除
  $ docker run -it \
      -v /home/someuser/registry:/registry \
      -e REGISTRY_URL=https://registry.example.com \
      -e SELF_SIGNED_CERT="true" \
      -e REGISTRY_AUTH="myuser:sickpassword" \
      mortensrasmussen/docker-registry-manifest-cleanup:1.1.1
  ```

> 若使用上面命令执行报错找不到目录的错误可切换`docker-registry-manifest-cleanup`的版本至1.0.5进行尝试

- 使用docker-registry-manifest-cleanup 1.0.5进行存储回收。
  ```console
  # 由于以前的版本不支持提权，故将 /etc/registry/config.yml 中的鉴权配置部分先暂时注释掉，重启registry容器
      # auth:
        # token:
          # issuer: harbor-token-issuer
          # realm: https://registry.example.com/service/token
          # rootcertbundle: /etc/registry/root.crt
          # service: harbor-registry

  # 执行以下脚本尝试通过api模拟删除manifests
  $ docker run -it --rm \
      -v /home/someuser/registry:/registry \
      -e REGISTRY_URL=https://registry.example.com \
      -e CURL_INSECURE=true \
      -e DRY_RUN=true \
      mortensrasmussen/docker-registry-manifest-cleanup:1.0.5
      
  # 如上一步没有报错，执行以下脚本，真正删除
  $ docker run -it --rm \
      -v /home/someuser/registry:/registry \
      -e REGISTRY_URL=https://registry.example.com \
      -e CURL_INSECURE=true \
      mortensrasmussen/docker-registry-manifest-cleanup:1.0.5

  # 执行完成后将授权配置改回来，取消注释
      auth:
        token:
          issuer: harbor-token-issuer
          realm: https://registry.example.com/service/token
          rootcertbundle: /etc/registry/root.crt
          service: harbor-registry
  ```

### 参考文档：
- https://github.com/mortensteenrasmussen/docker-registry-manifest-cleanup
- https://github.com/goharbor/harbor/blob/v1.7.0/docs/user_guide.md#deleting-repositories
