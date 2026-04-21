// Package access 提供门禁权限管理的HTTP接口处理器。
package access

import (
	"dac/entity/config"
	"fmt"
	"mime/multipart"
	"net/url"
	"strconv"

	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/logic/cgi/staff"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// GetAllStaffCompany 获取所有人员所属公司列表
func GetAllStaffCompany(c *gin.Context) {
	companies, err := staff.GetAllStaffCompany(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, companies)
}

// GetStaffs 分页查询人员列表，支持关键词搜索和公司过滤
func GetStaffs(c *gin.Context) {
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		ghttp.SendResponseWithError(c, fmt.Errorf("offset error: %w", err))
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		ghttp.SendResponseWithError(c, fmt.Errorf("limit error: %w", err))
		return
	}
	query := c.Query("query")
	company := c.Query("company")
	staffs, err := staff.GetStaffs(c, utils.GetMozuID(c), offset, limit, query, company)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, staffs)
}

// AddStaff 新增人员信息
func AddStaff(c *gin.Context) {
	var s db.Staff
	var err error
	if err = c.ShouldBind(&s); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var (
		data struct {
			ID db.IDType `json:"id"`
		}
	)
	if data.ID, err = staff.AddStaff(c, utils.GetMozuID(c), s); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, data)
}

// ImportStaffs 从Excel文件导入人员数据
func ImportStaffs(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var file *multipart.FileHeader = nil
	f := form.File["file"]
	if len(f) > 0 {
		file = f[0]
	}
	if file == nil {
		ghttp.SendResponseWithErrorMessage(c, fmt.Sprintf("file field is empty"))
		return
	}

	if err = staff.Import(c, utils.GetMozuID(c), file); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// ExportStaffs 导出人员数据为Excel文件
func ExportStaffs(c *gin.Context) {
	f, err := staff.Export(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=utf-8''%v",
		url.QueryEscape("人员信息.xlsx")))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write staff excel error: %v", err)
	}
}

// UpdateStaff 更新指定ID的人员信息
func UpdateStaff(c *gin.Context) {
	staffId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	var s db.Staff
	if err = c.ShouldBind(&s); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	err = staff.UpdateStaff(c, staffId, s, utils.GetMozuID(c))
	ghttp.SendResponseWithError(c, err)
}

// DeleteStaff 删除指定ID的人员，支持不同删除模式
func DeleteStaff(c *gin.Context) {
	staffId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	deleteMode := cgi.DeleteStaffOnly
	if queryParam := c.Query("mode"); len(queryParam) > 0 {
		n, err := strconv.Atoi(queryParam)
		if err != nil {
			ghttp.SendResponseWithError(c, err)
			return
		}
		deleteMode = cgi.DeleteStaffModeType(n)
	}

	if err = staff.DeleteStaff(c, staffId, deleteMode); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}
