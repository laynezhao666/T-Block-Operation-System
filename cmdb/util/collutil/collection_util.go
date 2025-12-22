package collutil

import "github.com/google/go-cmp/cmp"

var ignoreFieldCmpOption = cmp.FilterPath(func(path cmp.Path) bool {
	field := path.Last().String()
	return field == ".Id" || field == ".CreateAt" || field == ".UpdateAt"
}, cmp.Ignore())

// FindDiff 查找两个集合之间的差异,查找出新增的、删除的、不变的
//
//	@param oldColl  		旧的集合
//	@param newColl			新的集合
//	@param keyFunc  		集合中元素唯一标识生成函数
//	@return addColl []T		新增的元素
//	@return delColl []T		移除的元素
//	@return sameColl []T	不变的元素
func FindDiff[T any, R comparable](oldColl []T, newColl []T, keyFunc func(item T) R, deepEqual bool) (addColl []T, delColl []T, sameColl []T) {
	// 先将集合转化为map,方便查找比对
	oldCollMap := make(map[R]T)
	newCollMap := make(map[R]T)
	for _, item := range oldColl {
		oldCollMap[keyFunc(item)] = item
	}
	for _, item := range newColl {
		newCollMap[keyFunc(item)] = item
	}
	// 初始化结果
	addColl = make([]T, 0)
	delColl = make([]T, 0)
	sameColl = make([]T, 0)
	// 查找删除的元素
	for k, item := range oldCollMap {
		if _, ok := newCollMap[k]; !ok {
			delColl = append(delColl, item)
		}
	}
	// 查找新增的元素和不变的元素，这里遍历新列表而不是map，是为了新增的数据保持原来的顺序
	for _, newItem := range newColl {
		k := keyFunc(newItem)
		if oldItem, ok := oldCollMap[k]; !ok {
			addColl = append(addColl, newItem)
		} else {
			if deepEqual {
				if cmp.Equal(newItem, oldItem, ignoreFieldCmpOption) {
					sameColl = append(sameColl, newItem)
				} else {
					addColl = append(addColl, newItem)
					delColl = append(delColl, oldItem)
				}
			} else {
				sameColl = append(sameColl, newItem)
			}
		}
	}
	return addColl, delColl, sameColl
}
