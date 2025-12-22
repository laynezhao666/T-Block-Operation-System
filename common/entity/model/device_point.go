package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	PointCategoryUnknown    = 0 // 未知分类
	PointCategoryAllCollect = 1 // 全采集
	PointCategoryCollectStd = 2 // 采集+标准
	PointCategoryAllStd     = 3 // 全标准

	PointTypeConstant = 3
)

var (
	funcMap = map[string]struct{}{
		"max": {},
		"min": {},
		"sum": {},
		"avg": {},
		"abs": {},
		"MAX": {},
		"MIN": {},
		"SUM": {},
		"AVG": {},
		"ABS": {},
	}
)

// DevicePoint 设备测点信息
type DevicePoint struct {
	Id              int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                         // 主键ID
	DeviceGid       string    `gorm:"column:device_gid;type:varchar(255);comment:设备GID;not null;" json:"device_gid"`                                 // 设备GID
	DeviceNumber    string    `gorm:"column:device_number;type:varchar(255);comment:设备编码;not null;" json:"device_number"`                            // 设备编码
	BelongCollector string    `gorm:"column:belong_collector;type:varchar(255);comment:归属采集器;not null;" json:"belong_collector"`                     // 归属采集器
	PointNameEn     string    `gorm:"column:point_name_en;type:varchar(255);comment:测点名称英文;not null;" json:"point_name_en"`                          // 测点名称英文
	PointNameZh     string    `gorm:"column:point_name_zh;type:varchar(255);comment:测点名称中文;not null;" json:"point_name_zh"`                          // 测点名称中文
	PointKey        string    `gorm:"column:point_key;type:varchar(255);comment:测点标识;not null;" json:"point_key"`                                    // 测点标识
	PointCategory   int32     `gorm:"column:point_category;type:tinyint(4);comment:测点分类;not null;default:0;" json:"point_category"`                  // 测点分类
	PointRw         string    `gorm:"column:point_rw;type:varchar(16);comment:测点读写分类;not null;" json:"point_rw"`                                     // 测点读写分类
	PointLevel      string    `gorm:"column:point_level;type:varchar(16);comment:测点级别;not null;" json:"point_level"`                                 // 测点级别
	Expression      string    `gorm:"column:expression;type:text;comment:测点表达式;" json:"expression"`                                                  // 测点表达式
	ExpressionMap   string    `gorm:"column:expression_map;type:text;comment:测点映射(设备GID);" json:"expression_map"`                                    // 测点映射(设备GID)
	ExpressionMapZh string    `gorm:"column:expression_map_zh;type:text;comment:测点映射(设备编号);" json:"expression_map_zh"`                               // 测点映射(设备编号)
	ValueType       string    `gorm:"column:value_type;type:varchar(16);comment:测点值类型;not null;" json:"value_type"`                                  // 测点值类型
	ValueValidRange string    `gorm:"column:value_valid_range;type:varchar(255);comment:测点值有效范围;not null;" json:"value_valid_range"`                 // 测点值有效范围
	ValueUnit       string    `gorm:"column:value_unit;type:varchar(32);comment:测点值单位;not null;" json:"value_unit"`                                  // 测点值单位
	ValuePrecision  string    `gorm:"column:value_precision;type:varchar(16);comment:测点值精度;not null;" json:"value_precision"`                        // 测点值精度
	ValueEnum       string    `gorm:"column:value_enum;type:varchar(255);comment:值枚举映射;not null;" json:"value_enum"`                                 // 值枚举映射
	MozuId          int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                 // 所属模组ID
	CreateAt        time.Time `gorm:"<-:false;column:create_at;type:datetime;comment:记录创建时间;not null;default:CURRENT_TIMESTAMP;" json:"create_at"`   // 记录创建时间
	UpdateAt        time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 设备测点表名称
func (p *DevicePoint) TableName() string {
	return "t_device_point"
}

// CalcUniqueKey 计算设备测点唯一标识
func (p *DevicePoint) CalcUniqueKey() string {
	return fmt.Sprintf("%s|%s", p.DeviceGid, p.PointNameEn)
}

// CalcDependCollector 计算测点依赖的采集器
func (p *DevicePoint) CalcDependCollector(collectorToMajorCollectorMap map[string]string,
	collectorPointToStdPointMap map[string]string) *hashset.Set {
	result := hashset.New()
	relatePoints := strings.Split(strings.TrimSpace(strings.Trim(p.ExpressionMap, ";")), ";")
	relatePoints = lo.Filter(relatePoints, func(item string, index int) bool {
		return len(item) != 0
	})
	for _, point := range relatePoints {
		begin := strings.Index(point, "=")
		end := strings.Index(point, ".")
		deviceGid := point[begin+1 : end]
		pointId := point[begin+1:]
		result.Add(collectorToMajorCollectorMap[deviceGid])
		if len(relatePoints) == 1 {
			collectorPointToStdPointMap[pointId] = p.PointKey
		}
	}
	return result
}

// CollectorToStdPoint 用于将标准到采集的映射转化为标准到标准的映射
func (p *DevicePoint) CollectorToStdPoint(ctx context.Context, collectorPointToStdPointMap map[string]string) {
	relatePoints := strings.Split(strings.TrimSpace(strings.Trim(p.ExpressionMap, ";")), ";")
	newRefPoints := make([]string, 0, len(relatePoints))
	badRefPoints := make([]string, 0)
	for _, point := range relatePoints {
		begin := strings.Index(point, "=")
		pointId := point[begin+1:]
		if stdPointName, ok := collectorPointToStdPointMap[pointId]; ok {
			if stdPointName != p.PointKey {
				newRefPoints = append(newRefPoints, fmt.Sprintf("%s%s", point[:begin+1], stdPointName))
			} else {
				newRefPoints = append(newRefPoints, point)
			}
		} else {
			badRefPoints = append(badRefPoints, point)
		}
	}
	if len(badRefPoints) > 0 {
		log.WarnContextf(ctx, "bad std point [%s], ref point [%s] can not found",
			fmt.Sprintf("%s.%s", p.DeviceNumber, p.PointNameEn), strings.Join(badRefPoints, ";"))
	}
	p.ExpressionMap = strings.Join(newRefPoints, ";")
}

// FixGid 修正room类的设备GID
func (p *DevicePoint) FixGid() {
	if strings.HasPrefix(p.DeviceGid, "room") {
		p.DeviceGid = fmt.Sprintf("%d:%s", p.MozuId, p.DeviceGid)
		p.PointKey = fmt.Sprintf("%s.%s", p.DeviceGid, p.PointNameEn)
	}
}

// StdToCollectorPoint 用于将标准到标准的映射转化为标准到采集的映射
func (p *DevicePoint) StdToCollectorPoint(relateStdPoint map[string]*DevicePoint) error {
	if p.PointCategory == PointCategoryUnknown || p.PointCategory == PointCategoryAllCollect {
		return nil
	}
	// 解析出原始表达式及变量测点映射关系
	node, err := p.buildExpressNode("", relateStdPoint)
	if err != nil {
		return err
	}
	p.Expression = node.getExpress()
	p.ExpressionMap, p.ExpressionMapZh = node.getExpressMap()
	return nil
}

func getExpressMap(expression string, prefix string) map[string]string {
	trimExpress := strings.TrimSpace(strings.Trim(expression, ";"))
	if len(trimExpress) == 0 {
		return map[string]string{}
	}
	relatePoints := strings.Split(trimExpress, ";")
	return lo.SliceToMap(relatePoints, func(item string) (string, string) {
		eqPos := strings.Index(item, "=")
		return fmt.Sprintf("%s%s", prefix, item[:eqPos]), item[eqPos+1:]
	})
}

func (p *DevicePoint) buildExpressNode(parentVar string, stdPointMap map[string]*DevicePoint) (*expressNode, error) {
	tree, err := parser.Parse(p.Expression)
	if err != nil {
		return nil, fmt.Errorf("point:[%s], invalid expression [%s]", p.PointKey, p.Expression)
	}
	ast.Walk(&tree.Node, &modifyVarVisitor{parentVar: parentVar})
	varMapEn := getExpressMap(p.ExpressionMap, parentVar)
	varMapCn := getExpressMap(p.ExpressionMapZh, parentVar)
	children := make(map[string]*expressNode)
	leftVarMapEn := make(map[string]string)
	leftVarMapCn := make(map[string]string)
	for varName, childPointKey := range varMapEn {
		if childPoint, ok := stdPointMap[childPointKey]; ok {
			childNode, err := childPoint.buildExpressNode(varName, stdPointMap)
			if err != nil {
				return nil, err
			}
			children[varName] = childNode
		} else {
			// 已经是采集点，需要保存
			leftVarMapEn[varName] = childPointKey
			leftVarMapCn[varName] = varMapCn[varName]
		}
	}
	// 没有子节点，无需再扩展
	if len(children) == 0 {
		return &expressNode{
			node:     &tree.Node,
			varMapEn: leftVarMapEn,
			varMapCn: leftVarMapCn,
		}, nil
	}
	res := &expandVarVisitor{
		children:    children,
		newVarMapEn: leftVarMapEn,
		newVarMapCn: leftVarMapCn,
	}
	ast.Walk(&tree.Node, res)
	return &expressNode{
		node:     &tree.Node,
		varMapEn: res.newVarMapEn,
		varMapCn: res.newVarMapCn,
	}, nil
}

// 表达式节点，存储表达式ast树及变量映射
type expressNode struct {
	node     *ast.Node
	varMapEn map[string]string
	varMapCn map[string]string
}

// 获取ats树代表的表达式
func (obj *expressNode) getExpress() string {
	express := (*obj.node).String()
	express = strings.ReplaceAll(express, " ", "")
	return express
}

// 获取表达式中所有的变量及对应的测点
func (obj *expressNode) getExpressMap() (string, string) {
	pointKeyEns := make([]string, 0, len(obj.varMapEn))
	pointKeyCns := make([]string, 0, len(obj.varMapCn))
	existVarMap := make(map[string]struct{})
	res := &extraVarVisitor{}
	ast.Walk(obj.node, res)
	for _, varName := range res.vars {
		if _, ok := existVarMap[varName]; ok {
			continue
		}
		if pointKeyEn, ok := obj.varMapEn[varName]; ok {
			pointKeyEns = append(pointKeyEns, fmt.Sprintf("%s=%s", varName, pointKeyEn))
		}
		if pointKeyCn, ok := obj.varMapCn[varName]; ok {
			pointKeyCns = append(pointKeyCns, fmt.Sprintf("%s=%s", varName, pointKeyCn))
		}
		existVarMap[varName] = struct{}{}
	}
	return strings.Join(pointKeyEns, ";"), strings.Join(pointKeyCns, ";")
}

// 用于修改表达式中的变量
type modifyVarVisitor struct {
	parentVar string
}

func (obj *modifyVarVisitor) Visit(n *ast.Node) {
	if ident, ok := (*n).(*ast.IdentifierNode); ok {
		if _, ok := funcMap[ident.Value]; ok {
			return
		}
		ident.Value = fmt.Sprintf("%s%s", obj.parentVar, ident.Value)
	}
}

// 用于扩展表达式中的变量
type expandVarVisitor struct {
	children    map[string]*expressNode
	newVarMapEn map[string]string
	newVarMapCn map[string]string
}

func (obj *expandVarVisitor) Visit(n *ast.Node) {
	if ident, ok := (*n).(*ast.IdentifierNode); ok {
		if _, ok := funcMap[ident.Value]; ok {
			return
		}
		varName := ident.Value
		if childNode, ok := obj.children[varName]; ok {
			*n = *childNode.node
			obj.newVarMapEn = lo.Assign(obj.newVarMapEn, childNode.varMapEn)
			obj.newVarMapCn = lo.Assign(obj.newVarMapCn, childNode.varMapCn)
		}
	}
}

// 用于获取表达式中所有的变量
type extraVarVisitor struct {
	vars []string
}

func (obj *extraVarVisitor) Visit(n *ast.Node) {
	if ident, ok := (*n).(*ast.IdentifierNode); ok {
		if _, ok := funcMap[ident.Value]; ok {
			return
		}
		obj.vars = append(obj.vars, ident.Value)
	}
}
