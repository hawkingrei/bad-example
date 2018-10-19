package ecode

// ep ecode interval is [0,990000]
var (
	//merlin/paas
	MerlinInvalidClusterErr = New(80001) // 集群不合法
	MerlinPaasRequestErr    = New(80002) // Paas 请求错误

	//merlin/tree
	MerlinGetUserTreeFailed        = New(80010) //获取 tree 节点失败
	MerlinTreeResponseErr          = New(80011) //请求 Tree 失败
	MerlinShouldTreeFullPath       = New(80012) //节点不合法
	MerlinTreeRequestErr           = New(80013) //服务树请求错误
	MerlinLoseTreeContainerNodeErr = New(80014) //当前节点下不存在 dev/container 子节点

	//merlin
	MerlinDuplicateMachineNameErr = New(81001) // 机器名称重复
	MerlinInvalidMachineAmountErr = New(81002) //机器数量不合法
	MerlinInvalidNodeAmountErr    = New(81003) //挂载节点数必须大于0且不大于10
	MerlinUpdateNodeErr           = New(81004) //更新节点失败

	//other
	MerlinIllegalPageNumErr  = New(89001) //分页页码不合法
	MerlinIllegalPageSizeErr = New(89002) //分页大小不合法

	MerlinDelayMachineErr                     = New(89010) //机器自主延期失败
	MerlinApplyMachineErr                     = New(89011) //机器申请延期失败
	MerlinCancelMachineErr                    = New(89012) //机器取消延期失败
	MerlinAuditMachineErr                     = New(89013) //机器审核延期失败
	MerlinApplyMachineByApplyEndTimeMore3MErr = New(89014) //机器申请延期失败

	//hubbili
	MerlinHubRequestErr           = New(89015) //请求bilihub失败
	MerlinHubNoRight              = New(89016) //没有权限执行
	MerlinImagePullErr            = New(89017) //下载镜像失败
	MerlinImagePushErr            = New(89018) //上传镜像失败
	MerlinImageTagErr             = New(89019) //Tag镜像失败
	MerlinSnapshotInDoingErr      = New(89024) //快照进行中
	MerlinNoHubAccount            = New(89026) //该用户没有Hub账号
	MerlinDuplicateImageNameErr   = New(89028) //镜像名称重复
	MerlinMachine2ImageInDoingErr = New(89029) //机器转镜像进行中
	MerlinMachineImageNotSameErr  = New(89030) //镜像名称不一致

	MerlinDeviceNotBind              = New(89020) //设备未绑定
	MerlinDeviceFarmErr              = New(89021) //DeviceFarm Error
	MerlinDeviceFarmMachineStatusErr = New(89025) //Merlin Device Farm Machine StatusErr

	FootmanBuglyErr = New(89022) //FootmanBuglyErr
	FootmanTapdErr  = New(89023) //Tapd请求错误

	//melloi/PaaS
	MelloiPaasRequestErr     = New(60002) // Paas 请求错误
	MeilloiIllegalPageNumErr = New(60004) //分页页码不合法
	MeilloillegalPageSizeErr = New(60005) //分页大小不合法
	MelloiTreeRequestErr     = New(60001) // Tree 请求错误
	MelloiAdminExist         = New(60003)
)
