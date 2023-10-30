package pkg

import "net/http"

type (
	Code struct {
		Status   bool   `json:"status"`
		Code     int    `json:"code"`
		Message  string `json:"message"`
		HttpCode int    `json:"-"`
	}

	Response struct {
		*Code
		Data interface{} `json:"data,omitempty"`
	}
)

func makeCode(httpCode, code int, status bool, msg string) *Code {
	return &Code{
		Status:   status,
		Code:     code,
		Message:  msg,
		HttpCode: httpCode,
	}
}

func (c *Code) SetMsg(msg string) *Code {
	newCode := new(Code)
	newCode.Status = c.Status
	newCode.Code = c.Code
	newCode.HttpCode = c.HttpCode
	newCode.Message = msg
	return newCode
}

func MakeResp(code *Code, data interface{}) *Response {
	return &Response{
		Code: code,
		Data: data,
	}
}

var (
	Success = makeCode(http.StatusOK, 200, true, "success")
	Created = makeCode(http.StatusCreated, 200, true, "created")

	NotModify = makeCode(http.StatusNotModified, 0, true, "当前已是最新版本")

	NonceIsExpired        = makeCode(http.StatusBadRequest, 400101, false, "nonce is expired")
	InvalidInputSignature = makeCode(http.StatusBadRequest, 400102, false, "invalid input signature")
	InternalServerError   = makeCode(http.StatusBadRequest, 400103, false, "internal server error")
	LoginError            = makeCode(http.StatusBadRequest, 400104, false, "not logged in")
	AddrEmpty             = makeCode(http.StatusBadRequest, 400105, false, "地址不能为空")

	AppidOrSecretKeyEmpty        = makeCode(http.StatusBadRequest, 400001, false, "appid 或 secretKey 为空")
	ParamsEmpty                  = makeCode(http.StatusBadRequest, 400002, false, "必填项不能为空")
	InvalidMobile                = makeCode(http.StatusBadRequest, 400003, false, "手机号不合法")
	RepeatedOperate              = makeCode(http.StatusBadRequest, 400004, false, "操作过于频繁, 请稍后再试")
	AppidOrSecretKeyIncorrect    = makeCode(http.StatusBadRequest, 400005, false, "appid或secretKey不正确")
	ParamsError                  = makeCode(http.StatusBadRequest, 400006, false, "参数不合法")
	ParamsErrorI18n              = makeCode(http.StatusBadRequest, 400006, false, "Invalid parameter")
	ParamsCharacterError         = makeCode(http.StatusBadRequest, 400075, false, "密码格式不正确，包含特殊字符")
	VerifyCodeIncorrectOrExpired = makeCode(http.StatusBadRequest, 400007, false, "验证码不正确或已过期")
	UsernameExisted              = makeCode(http.StatusBadRequest, 400008, false, "用户名已存在")
	MobileExisted                = makeCode(http.StatusBadRequest, 400009, false, "手机号码已存在")
	EmailExisted                 = makeCode(http.StatusBadRequest, 400010, false, "邮箱已存在")
	EmailExistedI18n             = makeCode(http.StatusBadRequest, 400010, false, "Email already exists")
	InvalidUsername              = makeCode(http.StatusBadRequest, 400011, false, "用户名不合法")
	InvalidPassword              = makeCode(http.StatusBadRequest, 400012, false, "密码不合法")
	SamePassword                 = makeCode(http.StatusBadRequest, 4000121, false, "支付密码不能与登陆密码相同")
	InvalidEmail                 = makeCode(http.StatusBadRequest, 400013, false, "email不合法")
	InvalidEmailI18n             = makeCode(http.StatusBadRequest, 400013, false, "email illegal")
	UnsupportedTransferType      = makeCode(http.StatusBadRequest, 400014, false, "不支持的交易类型")
	BalanceOut                   = makeCode(http.StatusBadRequest, 400015, false, "余额不足")
	ContractNotSupported         = makeCode(http.StatusBadRequest, 400016, false, "不支持的合约")
	PositionNotExist             = makeCode(http.StatusBadRequest, 400017, false, "仓位不存在或已完成平仓")
	PositionOut                  = makeCode(http.StatusBadRequest, 400018, false, "仓位可平数量不足")
	UnknownOrderType             = makeCode(http.StatusBadRequest, 400019, false, "订单类型错误")
	OrderNotExist                = makeCode(http.StatusBadRequest, 400020, false, "订单不存在或已被删除")
	OrderCanceledOrFinished      = makeCode(http.StatusBadRequest, 400021, false, "订单已撤销或已完成交易")
	FileTypeUnsupported          = makeCode(http.StatusBadRequest, 400022, false, "不支持的文件类型")
	OrderQuantityOut             = makeCode(http.StatusBadRequest, 400023, false, "单笔订单数量不能超过500手")
	PositionQuantityOut          = makeCode(http.StatusBadRequest, 400024, false, "持仓总量不能超过2000手")
	UserNotExist                 = makeCode(http.StatusBadRequest, 400025, false, "用户不存在")
	ResourceNotExist             = makeCode(http.StatusBadRequest, 400026, false, "资源不存在或已被删除")
	RegisterCodeNotExist         = makeCode(http.StatusBadRequest, 400027, false, "注册码不存在")
	MustOnePrice                 = makeCode(http.StatusBadRequest, 400028, false, "必须至少有一个止盈/止损价格")
	ProfitPriceSet               = makeCode(http.StatusBadRequest, 400029, false, "已设置止盈价格, 无需重复设置")
	LossPriceSet                 = makeCode(http.StatusBadRequest, 400029, false, "已设置止损价格, 无需重复设置")
	BalanceWithdrawTooLittle     = makeCode(http.StatusBadRequest, 400029, false, "提币数量不能小于50")
	WithdrawAddrIncorrect        = makeCode(http.StatusBadRequest, 400030, false, "提币地址不正确, 必须是42位字符, 清检查")
	ProfitPriceEnterLongInvalid  = makeCode(http.StatusBadRequest, 400031, false, "开多单止盈价格必须大于开单价格")
	ProfitPriceEnterShortInvalid = makeCode(http.StatusBadRequest, 400032, false, "开空单止盈价格必须小于开单价格")
	LossPriceEnterLongInvalid    = makeCode(http.StatusBadRequest, 400033, false, "开多单止损价格必须小于开单价格")
	LossPriceEnterShortInvalid   = makeCode(http.StatusBadRequest, 400034, false, "开空单止损价格必须大于开单价格")
	VerifyInfoExisted            = makeCode(http.StatusBadRequest, 400035, false, "认证信息已提交或正在审核中")
	AddressBalanceEmpty          = makeCode(http.StatusBadRequest, 400036, false, "帐户无可用金额回笼")
	ProductNotExistOrUnMarket    = makeCode(http.StatusBadRequest, 400037, false, "商品不存在或暂未上市")
	PaymentMethodUnsupported     = makeCode(http.StatusBadRequest, 400038, false, "不支持的支付方式")
	TimePackageOut               = makeCode(http.StatusBadRequest, 400039, false, "托管包不足")
	ThereIsNoAvailableNode       = makeCode(http.StatusBadRequest, 400040, false, "当前账户无可用矿机")
	AssetIsNotEnough             = makeCode(http.StatusBadRequest, 400041, false, "账户余额不足")
	OrderStatusError             = makeCode(http.StatusBadRequest, 400042, false, "订单状态错误")
	AddressError                 = makeCode(http.StatusBadRequest, 400043, false, "充值地址或银行地址获取失败")
	UserLevelNotExist            = makeCode(http.StatusBadRequest, 400044, false, "获取用户等级失败")
	PoolNotExist                 = makeCode(http.StatusBadRequest, 400045, false, "矿池不存在")
	AddrNotExist                 = makeCode(http.StatusBadRequest, 400046, false, "地址不存在")
	NodePledgeStatusInvalid      = makeCode(http.StatusBadRequest, 400050, false, "矿机编号或抵押状态错误")
	NodePledgedTimeInvalid       = makeCode(http.StatusBadRequest, 400051, false, "矿机抵押未满两个月,不能解除抵押")
	AvailableAmountsInvalid      = makeCode(http.StatusBadRequest, 400061, false, "资产不足")
	WithdrawFail                 = makeCode(http.StatusBadRequest, 400062, false, "提现失败")
	WithdrawMinLimit             = makeCode(http.StatusBadRequest, 400063, false, "最小提现限制")
	WithdrawStatusError          = makeCode(http.StatusBadRequest, 400064, false, "提现状态错误")
	VerifyWithdrawError          = makeCode(http.StatusBadRequest, 400065, false, "审核提现失败")
	VerifiedWithdraw             = makeCode(http.StatusBadRequest, 400066, false, "已审核的提现")
	FundsPwdIncorrect            = makeCode(http.StatusBadRequest, 400067, false, "资金密码不正确")
	InvalidAddress               = makeCode(http.StatusBadRequest, 400068, false, "地址无效,请检查")
	UserRoleNotExist             = makeCode(http.StatusBadRequest, 400069, false, "获取用户role失败")
	UpdateUserRoleFail           = makeCode(http.StatusBadRequest, 400070, false, "修改用户role失败")

	PoolApplyCheckStatusError = makeCode(http.StatusBadRequest, 400080, false, "审核状态错误")
	PoolApplyStatusError      = makeCode(http.StatusBadRequest, 400080, false, "申请不是待审核状态")
	PoolApplyQualError        = makeCode(http.StatusBadRequest, 400080, false, "未满足申请条件")
	PoolApplySuccessError     = makeCode(http.StatusBadRequest, 400080, false, "请不要重复提交申请")

	TokenEmpty                  = makeCode(http.StatusUnauthorized, 401001, false, "未登录")
	UsernameOrPasswordIncorrect = makeCode(http.StatusUnauthorized, 401002, false, "用户名或密码不正确")
	LoginExpired                = makeCode(http.StatusUnauthorized, 401003, false, "登录信息已过期或未登录")
	PasswordIncorrect           = makeCode(http.StatusUnauthorized, 401004, false, "密码不正确")
	LoginOther                  = makeCode(http.StatusFailedDependency, 401005, false, "账号在其他的地方登陆")
	PermissionIncorrect         = makeCode(http.StatusUnauthorized, 401006, false, "用户权限不足")
	RepeatedSubmit              = makeCode(http.StatusUnauthorized, 401008, false, "请勿重复提交")
	FileNotFound                = makeCode(http.StatusNotFound, 404000, false, "删除地址不存在")

	FileIsTooLarge = makeCode(http.StatusRequestEntityTooLarge, 413000, false, "上传文件不能超过100MB")

	InternalError     = makeCode(http.StatusInternalServerError, 500001, false, "未知错误, 请稍后再试")
	InternalErrorI18n = makeCode(http.StatusInternalServerError, 500001, false, "Unknown error, please try again")

	RateAmountError     = makeCode(http.StatusBadRequest, 400071, false, "币价不能小于0")
	RateScopeError      = makeCode(http.StatusBadRequest, 400071, false, "涨跌幅上限不能低于下限")
	NumberTooLargeError = makeCode(http.StatusBadRequest, 400071, false, "数字太大")
	StockNo             = makeCode(http.StatusBadRequest, 400072, false, "库存不足")
	PhoneNoExist        = makeCode(http.StatusBadRequest, 400073, false, "手机号不存在")

	PromoCodeNoExist                   = makeCode(http.StatusBadRequest, 400100, false, "优惠码不存在")
	PromoCodeNoDesignatedGood          = makeCode(http.StatusBadRequest, 400101, false, "不是优惠码使用指定的商品")
	PromoCodeExpired                   = makeCode(http.StatusBadRequest, 400102, false, "优惠码已过期")
	PromoCodeNoUsageTimes              = makeCode(http.StatusBadRequest, 400103, false, "优惠码使用次数不足")
	PromoCodeUseByNewUser              = makeCode(http.StatusBadRequest, 400104, false, "优惠码只限制新用户使用")
	PromoCodeUserCreateTimeNoIn        = makeCode(http.StatusBadRequest, 400106, false, "用户注册时间不在优惠码指定的时间段")
	PromoCodeNoInPurchaseQuantityLimit = makeCode(http.StatusBadRequest, 400107, false, "超过优惠码使用的购买份数限制")
	PromoCodeNoUserUsageTime           = makeCode(http.StatusBadRequest, 400108, false, "超过优惠码使用的购买份数限制")
	PromoCodeNoServer                  = makeCode(http.StatusBadRequest, 400109, false, "优惠码已停用")
	PromoCodeIsExist                   = makeCode(http.StatusBadRequest, 400110, false, "优惠码已存在")
	PromoCodeQrcodeInvalid             = makeCode(http.StatusBadRequest, 400111, false, "无效二维码")
	PromoCodeQrcodeNoExist             = makeCode(http.StatusBadRequest, 400112, false, "二维码不存在")

	CheckCardBelong = makeCode(http.StatusBadRequest, 400201, false, "检查卡片是否属于此系列")

	MenuNoExist      = makeCode(http.StatusBadRequest, 400301, false, "权限菜单参数不合法")
	NoMenuPermission = makeCode(http.StatusBadRequest, 400302, false, "没有此菜单使用权限")
	ExcelIsNo        = makeCode(http.StatusBadRequest, 400303, false, "excels参数不合法")
	NONEWST          = makeCode(http.StatusBadRequest, 400305, false, "no news")

	RedPacketTxPending  = makeCode(http.StatusOK, 400401, false, "红包处于Pending状态，请等待上链")
	RedPacketOutOfSock  = makeCode(http.StatusBadRequest, 400402, false, "红包已被领完")
	RedPacketWasExpired = makeCode(http.StatusBadRequest, 400403, false, "红包已过期")

	AuthenticationEmpty   = makeCode(http.StatusBadRequest, 400500, false, "鉴权信息为空")
	AuthenticationInvalid = makeCode(http.StatusBadRequest, 400500, false, "鉴权未通过")
)
