package alerting

import (
	"github.com/datadefeat/datav/backend/internal/acl"
	"github.com/datadefeat/datav/backend/pkg/i18n"
	"github.com/datadefeat/datav/backend/pkg/common"
	"github.com/datadefeat/datav/backend/internal/session"
	"github.com/datadefeat/datav/backend/pkg/db"
	"github.com/datadefeat/datav/backend/pkg/models"
	"github.com/datadefeat/datav/backend/pkg/utils/simplejson"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func AddNotification(c *gin.Context) {
	nf  := &models.AlertNotification{}
	c.Bind(&nf)

	teamId,_ := strconv.ParseInt(c.Param("teamId"),10,64)
	if teamId == 0 {
		c.JSON(400, common.ResponseI18nError(i18n.BadRequestData))
		return 
	}

	if !acl.IsTeamEditor(teamId,c) {
		c.JSON(403, common.ResponseI18nError(i18n.NoPermission))
		return
	}

	settings,_ := nf.Settings.Encode()
	now := time.Now()
	_,err := db.SQL.Exec(`INSERT INTO alert_notification (team_id, name, type, is_default, disable_resolve_message, send_reminder, upload_image, settings, created_by, created, updated) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
	teamId, nf.Name, nf.Type, nf.IsDefault, nf.DisableResolveMessage, nf.SendReminder, nf.UploadImage, settings, session.CurrentUserId(c), now, now)
	if err != nil {
		logger.Warn("add alert notification error", "error", err)
		c.JSON(500, common.ResponseInternalError())
		return 
	}

}

func UpdateNotification(c *gin.Context) {
	nf  := &models.AlertNotification{}
	c.Bind(&nf)

	teamId,_ := strconv.ParseInt(c.Param("teamId"),10,64)
	if nf.Id == 0 || teamId == 0 || teamId != nf.TeamId {
		c.JSON(400, common.ResponseI18nError(i18n.BadRequestData))
		return 
	}

	if !acl.IsTeamEditor(teamId,c) {
		c.JSON(403, common.ResponseI18nError(i18n.NoPermission))
		return
	}

	settings,_ := nf.Settings.Encode()
	now := time.Now()
	_,err := db.SQL.Exec(`UPDATE alert_notification SET name=?, type=?, is_default=?, disable_resolve_message=?, send_reminder=?, upload_image=?, settings=?, updated=? WHERE id=?`,
	nf.Name, nf.Type, nf.IsDefault, nf.DisableResolveMessage, nf.SendReminder, nf.UploadImage, settings, now, nf.Id)
	if err != nil {
		logger.Warn("add alert notification error", "error", err)
		c.JSON(500, common.ResponseInternalError())
		return 
	}
}

func DeleteNotification(c *gin.Context) {
	id,_ := strconv.ParseInt(c.Param("id"),10,64)
	if id == 0 {
		c.JSON(400, common.ResponseI18nError(i18n.BadRequestData))
		return 
	}

	notification,err := models.QueryNotification(id)
	if err != nil {
		logger.Warn("query notification error","error",err)
		c.JSON(400, common.ResponseInternalError())
		return 
	}

	if !acl.IsTeamEditor(notification.TeamId,c) {
		c.JSON(403, common.ResponseI18nError(i18n.NoPermission))
		return
	}

	_,err = db.SQL.Exec(`DELETE FROM alert_notification WHERE id=?`,id)
	if err != nil {
		logger.Warn("get alert notification error", "error", err)
		c.JSON(500, common.ResponseInternalError())
		return 
	}
}

func GetNotifications(c *gin.Context) {
	teamId,_ := strconv.ParseInt(c.Param("teamId"),10,64)
	if teamId == 0 {
		c.JSON(400, common.ResponseI18nError(i18n.BadRequestData))
		return 
	}

	rows,err := db.SQL.Query(`SELECT id,name,type,is_default, disable_resolve_message, send_reminder, upload_image, settings FROM alert_notification WHERE team_id=?`,teamId)
	if err !=nil {
		logger.Warn("get alert notification error", "error", err)
		c.JSON(500, common.ResponseInternalError())
		return 
	}

	notifications := make([]*models.AlertNotification,0)
	for rows.Next() {
		n := &models.AlertNotification{}
		var rawSetting []byte
		err := rows.Scan(&n.Id,&n.Name,&n.Type,&n.IsDefault,&n.DisableResolveMessage,&n.SendReminder,&n.UploadImage,&rawSetting)
		if err != nil {
			logger.Warn("scan alerting notification error", "error", err)
			continue
		}

		setting := simplejson.New()
		err = setting.UnmarshalJSON(rawSetting)
		if err != nil {
			logger.Warn("unmarshal alerting notification setting error", "error", err)
			continue
		}

		n.Settings = setting
		n.TeamId = teamId

		notifications = append(notifications, n)
	}

	c.JSON(200, common.ResponseSuccess(notifications))
}