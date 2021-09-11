package utils

import "swan-miner/logs"

const HTTP_CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
const HTTP_CONTENT_TYPE_JSON = "application/json; charset=utf-8"

const GET_OFFLINEDEAL_LIMIT_DEFAULT = 50

var logger = logs.GetLogger()
