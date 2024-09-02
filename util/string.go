package util

// 判断字母是不是大写
func IsASCIIUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

// 字母大小写转换
func UpperLowerExchange(c byte) byte {
	return c ^ ' ' // 空格是32 小写字母a是97 大写字母A是65 相差32 大写字母和小写字母只是二进制的第六不相同
}

// 把驼峰命名改成蛇形命名法 camelCase -> camel_case
func Camel2Snake(s string) string {
	if len(s) == 0 {
		return ""
	}
	t := make([]byte, 0, len(s)+4) // 预留_的位置
	// 判断首个字母
	if IsASCIIUpper(s[0]) {
		t = append(t, UpperLowerExchange(s[0]))
	} else {
		t = append(t, s[0])
	}
	// 遍历其他字母
	for i := 1; i < len(s); i++ {
		if IsASCIIUpper(s[i]) {
			t = append(t, '_', UpperLowerExchange(s[i]))
		} else {
			t = append(t, s[i])
		}
	}

	return string(t)
}
