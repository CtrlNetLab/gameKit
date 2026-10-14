--[[
-- usage: checkOnline.lua keyname gid uinxtime plattype userid state accid
-- keyname: 表名，名字一般为 online[_{who}]:plattype:userid or online[_{who}]:accid
-- gid: gamesvr id
-- unixtime: unix 时间戳
-- plattype: 平台类型
-- userid: 用户 id
-- state: 登陆状态
-- accid: 账号id
--
-- hashtable online_derrick:plat:userid gid time plat uid state
-- 情况-0: 参数个数错误，返回 0
-- 情况-1: 没有数据，则直接设置，返回 1
-- 情况-2: 有数据，判断旧的 gid 是否跟新的一样，一样则返回 2, 且更新 time/state
-- 情况-3: 有数据，判断旧的 gid 是否跟新的一样，不一样则返回 3, gid, time, state
]]--

if #ARGV ~= 7 then
    return {0}
end

local plat = ARGV[4]
local userid = ARGV[5]
local gid = ARGV[2]
local state = ARGV[6]
local time = ARGV[3]
local accid = ARGV[7]
local userIdKey = ARGV[1] .. ":" .. plat .. ":" .. userid
local accIdKey = ARGV[1] .. ":" .. accid

local data = redis.call("hgetall", accIdKey)

redis.call("hmset", userIdKey, "gid", gid, "time", time, "plat", plat, "uid", userid, "state", state, "accid", accid)
redis.call("hmset", accIdKey, "gid", gid, "time", time, "plat", plat, "uid", userid, "state", state, "accid", accid)

-- 情况-1
if not data or #data <= 0 then
    return {1}
end

-- 情况-2
if gid == data[2] then
    return {2}
end

-- 情况-3
return {3, data[2], data[4], data[10]}

