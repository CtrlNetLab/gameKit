package universal

const (
	/**
	 * 数据块操作类型
	 */
	CHUNK_OPERATION_ALL     int32 = 1
	CHUNK_OPERATION_ADD     int32 = 2
	CHUNK_OPERATION_MODIFY  int32 = 3
	CHUNK_OPERATION_DELETE  int32 = 4
	CHUNK_OPERATION_REVERT  int32 = 5
	CHUNK_OPERATION_PUSH    int32 = 5
	CHUNK_OPERATION_NO_PUSH int32 = 6

	/**
	 * 推送客户端extend数据常量集合
	 */
	//账号数据块
	CHUNK_TYPE_ACCOUNT int32 = 0

	//基础数据块
	CHUNK_TYPE_BASIC int32 = 1

	//英雄数据块
	CHUNK_TYPE_HERO int32 = 2

	//背包（道具）数据块
	CHUNK_TYPE_ITEM int32 = 5

	//抽卡数据块
	CHUNK_TYPE_AGGREGATE_LOTTERY int32 = 10

	//布阵数据块
	CHUNK_TYPE_ARRAY int32 = 11

	//货币数据块
	CHUNK_TYPE_CURRENCY int32 = 12

	//关卡数据块
	CHUNK_TYPE_STAR_STAGE int32 = 13

	//章节任务数据块
	CHUNK_TYPE_TASK_CHAPTER int32 = 14

	//永久购买记录数据块
	CHUNK_TYPE_STORE_BUY_RECORD int32 = 15

	//每日商店礼包数据块
	CHUNK_TYPE_STORE_BAG int32 = 16

	//商店积分奖励数据块
	CHUNK_TYPE_STORE_POINT_AWARD int32 = 17

	//邮件数据块
	CHUNK_TYPE_MAIL_INFO int32 = 18

	//商城系统礼包数据块
	CHUNK_TYPE_SHOP_BAG int32 = 19

	//每日任务数据块dayFlags.taskDays.list
	CHUNK_TYPE_TASK_DAY int32 = 20

	//新手引导数据块
	CHUNK_NEW_GUIDE int32 = 21

	//地图解锁关卡数据块 mapMatter.unlockStage
	CHUNK_MAP_UNLOCK_STAGE int32 = 22

	//地图迷雾数据块 mapMatter.fogItems
	CHUNK_MAP_FOG int32 = 23

	//地图累计奖励数据块 mapMatter.stagesAwards
	CHUNK_MAP_STAGE_AWARD int32 = 24

	//地图最新关卡数据块 mapMatter.newestStage
	CHUNK_MAP_NEWEST_STAGE int32 = 25

	//每日任务数据块dayFlags.taskDays.info
	CHUNK_TYPE_TASK_DAY_INFO int32 = 26

	//每日任务数据块dayFlags.taskDays.list
	CHUNK_TYPE_TASK_DAY_LIST int32 = 27

	//章节任务数据块taskChapter.info
	CHUNK_TYPE_TASK_CHAPTER_INFO int32 = 28

	//章节任务数据块taskChapter.list
	CHUNK_TYPE_TASK_CHAPTER_LIST int32 = 29

	//临时英雄数据块
	CHUNK_TYPE_TEMP_HERO int32 = 30

	//PVP竞技场数据块 （已弃用：功能重做）
	// CHUNK_TYPE_PVP_ARENA int32 = 31

	// 消除秘境 -- 收集物资玩法
	CHUNK_TYPE_COLLECT_SUPPLIES int32 = 33

	// 消除秘境 -- 飞船维护玩法
	CHUNK_TYPE_SPACECRAFT_REPAIR int32 = 34

	// 消除秘境 -- 寻找泡泡玩法
	CHUNK_TYPE_TREASURE int32 = 35

	// 消除秘境 -- 保护泡泡玩法
	CHUNK_TYPE_BUBBLE int32 = 36

	// 爬塔副本
	CHUNK_TYPE_TOWER int32 = 37

	// 商城系统
	CHUNK_TYPE_SHOP_BUY int32 = 38

	// 每日爬塔类
	CHUNK_TYPE_TOWER_DAY int32 = 39

	// 商城货币消耗数据块
	CHUNK_TYPE_CONSUME_CURRENCY int32 = 40

	// 解锁新的功能
	CHUNK_TYPE_UNLOCK_SYSTEM int32 = 41

	// 消除秘境(全局块)
	CHUNK_TYPE_SECRET_ZONE int32 = 42

	// 抽卡累计全局统计块
	CHUNK_TYPE_AGGREGATE_LOTTERY_TOTAL int32 = 43

	// 地图障碍物数据块
	CHUNK_TYPE_MAP_EVENT int32 = 44

	// 英雄战力计算值数据块
	CHUNK_TYPE_HERO_COMBAT_VALUE int32 = 45

	// 每日领取在线奖励
	CHUNK_TYPE_ONLINE_AWARD int32 = 46

	// 签到模块子类的同步数据块
	CHUNK_TYPE_SIGN_IN_ITEM int32 = 47

	// 每日签到模块(dayFlags)的同步数据块
	CHUNK_TYPE_SIGN_IN_DAY_FLAGS int32 = 48

	// 签到模块的同步数据块
	CHUNK_TYPE_SIGN_IN int32 = 49

	//战令数据块同步标志
	CHUNK_TYPE_STAGE_BATTLE_ACTIVITY int32 = 50

	//主线任务数据块同步标志
	CHUNK_TYPE_TASK_MAIN int32 = 51
	//主线任务数据块taskMain.info
	CHUNK_TYPE_TASK_MAIN_INFO int32 = 52
	//主线任务数据块taskMain.list
	CHUNK_TYPE_TASK_MAIN_LIST int32 = 53

	//装备数据块
	CHUNK_TYPE_EQUIP_DATA int32 = 54

	//更新地图全局状态
	CHUNK_GLOBAL_MAP_STATE int32 = 55

	//派遣任务数据块
	CHUNK_TYPE_DISPATCH_TASK int32 = 56

	//派遣任务详情列表数据块
	CHUNK_TYPE_DISPATCH_TASK_DETAIL int32 = 57

	//流星宝箱数据块
	CHUNK_TYPE_METEOR int32 = 58

	//流星宝箱详情列表数据块
	CHUNK_TYPE_METEOR_BOX_DETAIL int32 = 59

	//流星宝箱每日观看广告次数
	CHUNK_TYPE_METEOR_AD_COUNT int32 = 60

	//登录信息数据块
	CHUNK_TYPE_LOGIN_INFO int32 = 61

	//第三方绑定数据块
	CHUNK_TYPE_PLATFORM_LIST int32 = 62

	//pvp重连
	CHUNK_TYPE_LIVING_PVP int32 = 63

	//回归角色的奖励Chunk
	CHUNK_TYPE_RETURNING_AWARD int32 = 64
	//地图关卡记录 数据块 mapMatter.stagesCounts
	CHUNK_MAP_STAGE_COUNT int32 = 65

	//PVP每日连胜数 (待废弃)
	CHUNK_PVP_WIN_TOTAL int32 = 66
	//PVP连胜奖励各阶段领取状态 (带废弃)
	CHUNK_PVP_SUCCESSIVE_FLAGS int32 = 67
	//活动数据
	CHUNK_TYPE_ACTIVITY_DATA int32 = 68
	//新手活动任务同步块 acitivtes.taskcommon
	CHUNK_TYPE_TASK_ACTIVITY_LIST int32 = 69

	//PVP布阵同步数据块
	CHUNK_PVP_ARRAY int32 = 70
	//PVP战斗记录同步块
	//CHUNK_PVP_CHALLENGE_RECORD int32 = 71
	//pvp数据块
	CHUNK_PVP_DATA int32 = 73

	//战令活动任务同步块taskcommon
	CHUNK_TYPE_STAGE_BATTLE_ACTIVITY_TASK int32 = 72
	//广告视频记录同步
	CHUNK_TYPE_Advertise_Video int32 = 74

	//pvp每日挑战次数
	//CHUNK_PVP_DAILY_CHALLENGE int32 = 75
	//pvp领取每日奖励(1.5版本，75合并到76)
	CHUNK_PVP_DAILY_REWARD int32 = 76
	//好友跨天重置块
	CHUNK_CROSS_DAY_FRIEND_DATA int32 = 77
	//pvp领取赛季段位奖励
	CHUNK_PVP_SEASON_DIVISION_REWARD int32 = 78

	// VIP过期时间
	CHUNK_VIP_EXPIRATION_TIME int32 = 79
	// VVIP过期时间
	CHUNK_VVIP_EXPIRATION_TIME int32 = 80
	// VIP获取每日奖励
	CHUNK_VIP_GET_DAY_REWARD int32 = 81
	// VVIP获取每日奖励
	CHUNK_VVIP_GET_DAY_REWARD int32 = 82
	//英雄历史配置
	CHUNK_HERO_HISTORY int32 = 83
	//评论
	CHUNK_HERO_COMMENT int32 = 84

	//卡片收集
	CHUNK_COLLECT int32 = 85
	//卡片列表数据
	CHUNK_COLLECT_CARD int32 = 86
	//卡片组数据
	CHUNK_COLLECT_GROUP int32 = 87
	// 秘境活动 -- 挖掘机
	CHUNK_TYPE_SECRET_ZONE_ACTIVITY_EXCAVATOR int32 = 88
	//英雄星级统计
	CHUNK_TYPE_HERO_STAR_STATISTICS int32 = 89
	//社区分享
	CHUNK_TYPE_LINK_SHARE int32 = 91
	//地图最外层数据块 mapMatter.LimitedTimeEffectEnergy
	CHUNK_MAP_MATTER_Effect_Energy int32 = 90
	//秘境活动 -- 宝藏挖掘
	CHUNK_TYPE_SECRET_ZONE_ACTIVITY_TREASURE int32 = 93
	// 消除秘境 -- 收集物资玩法(三消)
	CHUNK_TYPE_COLLECT_SUPPLIES_THREE int32 = 94
	//秘境活动 -- 轮船航行
	CHUNK_TYPE_SECRET_ZONE_ACTIVITY_NAVIGATION int32 = 95
	//活动结算后部分活动需要做动画得数据
	CHUNK_TYPE_ACTIVITY_RESULT_DATA int32 = 96
	//timeLine 同步块
	CHUNK_TYPE_TimeLine int32 = 97
	//邀请数据变动
	CHUNK_TYPE_Invite int32 = 98
	//秘境活动 -- 拼图
	CHUNK_TYPE_SECRET_ZONE_ACTIVITY_PUZZLE int32 = 99
	//飞船同步块 airshipInfo
	CHUNK_TYPE_AIRSHIP_INFO int32 = 100
	//英雄派遣状态
	CHUNK_TYPE_HERO_DISPATCH_STATUS int32 = 101
	//玩家shop数据
	CHUNK_TYPE_ACCOUNT_SHOP int32 = 102
	//英雄收集全局buff
	CHUNK_TYPE_GLOBAL_HERO_COLLECT = 103
	//活动排行榜结算后数据
	CHUNK_TYPE_ACTIVITY_RANK_RESULT_DATA int32 = 104
	// VIP数据块
	CHUNK_VIP_DATA int32 = 105
	// VVIP数据块
	CHUNK_VVIP_DATA int32 = 106
	//宝箱怪
	CHUNK_TYPE_CHESTMONSTER_DATA int32 = 107
	// VIP Time Reward数据块
	CHUNK_VIP_TIME_REWARD int32 = 108
	// VVIP Time Reward数据块
	CHUNK_VVIP_TIME_REWARD int32 = 109
	// 商城每日在线奖励
	CHUNK_TYPE_MALL_ONLINE_REAWARD int32 = 110
	//替补队伍数据块
	CHUNK_RESERVE_ARRAY int32 = 111
)
