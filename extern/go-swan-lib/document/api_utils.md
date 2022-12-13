# Groups
* [File](#File)
  * [IsFileExists](#IsFileExists)
  * [GetPathType](#GetPathType)
  * [RemoveFile](#RemoveFile)
  * [GetFileSize](#GetFileSize)
  * [GetFileSize2](#GetFileSize2)
  * [CopyFile](#CopyFile)
  * [CreateFileWithContents](#CreateFileWithContents)
  * [ReadAllLines](#ReadAllLines)
  * [ReadFile](#ReadFiles)
* [Json](#Json)
  * [GetFieldFromJson](#GetFieldFromJson)
  * [GetFieldStrFromJson](#GetFieldStrFromJson)
  * [GetFieldMapFromJson](#GetFieldMapFromJson)
  * [ToJson](#ToJson)
* [Common](#Common)
  * [GetInt64FromStr](#GetInt64FromStr)
  * [GetFloat64FromStr](#GetFloat64FromStr)
  * [GetIntFromStr](#GetIntFromStr)
  * [GetNumStrFromStr](#GetNumStrFromStr)
  * [GetByteSizeFromStr](#GetByteSizeFromStr)
  * [IsSameDay](#IsSameDay)
  * [GetRandInRange](#GetRandInRange)
  * [IsStrEmpty](#IsStrEmpty)
  * [GetDayNumFromEpoch](#GetDayNumFromEpoch)
  * [GetEpochFromDay](#GetEpochFromDay)
  * [GetMinFloat64](#GetMinFloat64)
  * [GetCurrentEpoch](#GetCurrentEpoch)
  * [GetDecimalFromStr](#GetDecimalFromStr)
  * [UrlJoin](#UrlJoin)

### IsFileExists

Inputs:
```shell
filePath
fileName string
```

Outputs:
```shell
bool
```
### GetPathType

Inputs:
```shell
dirFullPath string
```

Outputs:
```shell
int
```
### RemoveFile

Inputs:
```shell
filePath, fileName string
```

Outputs:
```shell
```
### GetFileSize

Inputs:
```shell
fileFullPath string
```

Outputs:
```shell
int64
```
### GetFileSize2

Inputs:
```shell
dir, fileName string
```

Outputs:
```shell
int64
```
### CopyFile

Inputs:
```shell
srcFilePath, destFilePath string
```

Outputs:
```shell
int64, error
```
### CreateFileWithContents

Inputs:
```shell
filepath string, lines []string
```

Outputs:
```shell
int, error
```
### ReadAllLines

Inputs:
```shell
dir, filename string
```

Outputs:
```shell
[]string, error
```
### ReadFile

Inputs:
```shell
filePath string
```

Outputs:
```shell
string, []byte, error
```
