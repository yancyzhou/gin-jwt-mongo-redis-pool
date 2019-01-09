package Api

import (
	"TimeLine/Lib"
	M "TimeLine/Model"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

// 登录参数
type PersonParam struct {
	NickName string `form:"nickname" json:"nickname" binding:"required"`
	Sex      int    `form:"sex" json:"sex" binding:"-"`
	Birthday string `form:"birthday" json:"birthday" binding:"required"`
	Born     int    `form:"born" json:"born" binding:"-"`
	Role     int    `form:"role" json:"role" binding:"-"`
	OpenId   string `form:"openid" json:"openid" binding:"required"`
}

func GetGrowthStandards(c *gin.Context) {
	//if _, err := Lib.Set(c, "testkey", "1234567890");err!=nil{
	//	c.AbortWithStatusJSON(http.StatusInternalServerError,gin.H{"code":http.StatusInternalServerError,
	//		"msg":http.StatusText(http.StatusInternalServerError)})
	//
	//}
	//if replyS, err := Lib.Get(c, "testkey");err!=nil{
	//	c.AbortWithStatusJSON(http.StatusInternalServerError,gin.H{"code":http.StatusInternalServerError,
	//		"msg":http.StatusText(http.StatusInternalServerError)})
	//}else {
	//	fmt.Println(replyS)
	//	//获取到 redis中的指定的 key的 value值 做对应的操作
	//}
	//claims, b := Lib.GetPayLoad(c)
	//payload := Lib.CustomClaims{}.Payload
	//if b {
	//	payload = claims.Payload
	//}
	skip, _ := strconv.Atoi(c.Param("skip"))
	limit, _ := strconv.Atoi(c.Param("limit"))
	gs := []M.GrowthStandard{}
	e := M.GrowthStandards().Find(nil).Sort("Type", "Days").Skip(skip).Limit(limit).All(&gs)

	if e != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
			"msg": http.StatusText(http.StatusInternalServerError)})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": gs, "Total": len(gs)})
	}

}

func CreatePerson(ctx *gin.Context) {
	var PersonP PersonParam
	if err := ctx.ShouldBindJSON(&PersonP); err == nil {
		result := M.Person{Id: bson.NewObjectId(), NickName: PersonP.NickName,
			Sex: PersonP.Sex, Birthday: PersonP.Birthday, Born: PersonP.Born, Role: PersonP.Role, CreateDateTime: time.Now()}
		err := M.Persons().Insert(&result)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusPaymentRequired, "msg": err.Error()})
		} else {
			userone := &M.User{}
			user := M.User{WxOpenId: PersonP.OpenId, PersonId: result.Id.Hex()}
			err := M.Users().Find(bson.M{"WxOpenId": PersonP.OpenId}).One(&userone)
			if err != nil {
				user.CreateDateTime = time.Now()
			}
			ups, err := M.Users().Upsert(bson.M{"WxOpenId": PersonP.OpenId}, user)
			if err != nil {
				defer M.Rollback(M.CollectionName_Person, result.Id)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": err.Error()})
			} else {
				ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": ups.UpsertedId})
			}
		}
	} else {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusPaymentRequired, "msg": err.Error()})
	}
}

//获取用户信息

type UserInfo struct {
	OpenId string `form:"openid" json:"openid" binding:"required"`
}

type UserResult struct {
	User M.Person
	Days int
}

func GetUserInfo(c *gin.Context) {
	var userinfo UserInfo
	if err := c.ShouldBindJSON(&userinfo); err == nil {
		userone := &M.User{}
		result := &UserResult{}
		err := M.Users().Find(bson.M{"WxOpenId": userinfo.OpenId}).One(&userone)
		if err != nil {
			result.User = M.Person{}
			result.Days = -1
		} else {
			loc, _ := time.LoadLocation("Local")
			M.Persons().FindId(bson.ObjectIdHex(userone.PersonId)).One(&result.User)
			toBeCharge := result.User.Birthday + " 00:00:00"
			parse_str_time, _ := time.ParseInLocation("2006-01-02 15:04:05", toBeCharge, loc)
			result.Days = Lib.TimeSub(time.Now(), parse_str_time)
		}
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": result})
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusPaymentRequired, "msg": err.Error()})
	}
}
