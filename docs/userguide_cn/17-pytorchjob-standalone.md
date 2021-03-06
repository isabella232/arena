
这个示例展示了如何使用 `Arena` 提交一个 pytorch 单机的作业。该示例将从 git url 下载源代码。

1. 第一步是检查可用的GPU资源
    ```
    ➜ arena top node
    NAME                       IPADDRESS     ROLE    STATUS  GPU(Total)  GPU(Allocated)
    cn-huhehaote.172.16.0.205  172.16.0.205  master  ready   0           0
    cn-huhehaote.172.16.0.206  172.16.0.206  master  ready   0           0
    cn-huhehaote.172.16.0.207  172.16.0.207  master  ready   0           0
    cn-huhehaote.172.16.0.208  172.16.0.208  <none>  ready   4           0
    cn-huhehaote.172.16.0.209  172.16.0.209  <none>  ready   4           0
    cn-huhehaote.172.16.0.210  172.16.0.210  <none>  ready   4           0
    -----------------------------------------------------------------------------------------
    Allocated/Total GPUs In Cluster:
    0/12 (0%)
    ```
    有 3 个包含 GPU 的可用节点用于运行训练作业

2. 提交一个 pytorch 训练作业，本示例从 [阿里云 code](https://code.aliyun.com/370272561/mnist-pytorch.git) 下载源代码
    ```
    # 单机单卡
    ➜ arena --loglevel info submit pytorch \
            --name=pytorch-local-git \
            --gpus=1 \
            --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
            --sync-mode=git \
            --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
            "python /root/code/mnist-pytorch/mnist.py --backend gloo"
    configmap/pytorch-local-git-pytorchjob created
    configmap/pytorch-local-git-pytorchjob labeled
    pytorchjob.kubeflow.org/pytorch-local-git created
    INFO[0000] The Job pytorch-local-git has been submitted successfully
    INFO[0000] You can run `arena get pytorch-local-git --type pytorchjob` to check the job status
    ```

    > 这会下载源代码，并将其解压缩到工作目录的 `code/` 目录。默认的工作目录是 `/root`，您也可以使用 `--workingDir` 加以指定。
    
    > 如果您正在使用非公开 git 代码库，则可以使用以下命令：

    ```
    ➜ arena --loglevel info submit pytorch \
        --name=pytorch-local-git \
        --gpus=1 \
        --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
        --sync-mode=git \
        --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
        --env=GIT_SYNC_USERNAME=yourname \
        --env=GIT_SYNC_PASSWORD=yourpwd \
        "python /root/code/mnist-pytorch/mnist.py --backend gloo"
    ```

3. 列出所有作业
    ```
    ➜ arena list
    NAME                          STATUS     TRAINER     AGE  NODE
    pytorch-local-git             SUCCEEDED  PYTORCHJOB  21h  N/A
    ```

4. 获取作业的详细信息
    ```
    ➜ arena get pytorch-local-git
    STATUS: SUCCEEDED
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 35s
    
    NAME               STATUS     TRAINER     AGE  INSTANCE                    NODE
    pytorch-local-git  SUCCEEDED  PYTORCHJOB  23h  pytorch-local-git-master-0  172.16.0.210
    ```

5. 检查日志
    ``` 
    ➜ arena logs pytorch-local-git
    WORLD_SIZE: 1, CURRENT_RANK: 0
    args: Namespace(backend='gloo', batch_size=64, data='/root/code/mnist-pytorch', dir='/root/code/mnist-pytorch/logs', epochs=1, log_interval=10, lr=0.01, momentum=0.5, no_cuda=False, save_model=False, seed=1, test_batch_size=1000)
    Using CUDA
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:541: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint8 = np.dtype([("qint8", np.int8, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:542: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint8 = np.dtype([("quint8", np.uint8, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:543: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint16 = np.dtype([("qint16", np.int16, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:544: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint16 = np.dtype([("quint16", np.uint16, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:545: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint32 = np.dtype([("qint32", np.int32, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:550: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      np_resource = np.dtype([("resource", np.ubyte, 1)])
    Train Epoch: 1 [0/60000 (0%)]	loss=2.3000
    Train Epoch: 1 [640/60000 (1%)]	loss=2.2135
    Train Epoch: 1 [1280/60000 (2%)]	loss=2.1705
    Train Epoch: 1 [1920/60000 (3%)]	loss=2.0767
    Train Epoch: 1 [2560/60000 (4%)]	loss=1.8681
    ...
    ```
