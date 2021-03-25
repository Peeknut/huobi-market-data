package risk_manage

import "sort"

const (
	SELL_IOC = "sell-ioc"
	BUY_LIMIT = "buy-limit"
)

var (
	accountHold [2][4]float64
	profit	[2][2]float64
	confusion_list [2][2]int
	trade_price float64
	signals_risk_mng = [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}
	signals_risk_action [4]int = [4]int{0,0,0,0}
	model_pred_data bool
	is_pred_data_change [2]int

	instrumentHold map[string]*InstrumentHoldValue
	remainHold map[string]*InstrumentHoldValue
	orderBuy map[string]*OrderBuyValue
	orderSell map[string]*OrderSellValue
)

type InstrumentHoldValue struct {
	ClientOrderId   string   //客户端订单ID
	OfficialOrderId int64    //交易所订单ID
	VolHold         float64  //持币数量
	CostHold        float64  //持币成本（买入价格）
	FreezeSize      float64  //冻结数量
	LineStopLoss    float64  //止损线
	VolStopLoss     float64  //止损量
	BuyFund         float64  //买入资金
	SellFund        float64  //卖出资金
	OrderStatus     int8     //订单状态
	TypeStopLoss    int8     //止损线类型
	TradeNum        int32    //成交次数
	OrderStatus2    int8     //止损线类型2
	SplitOrder      []string //拆单ID
	CombinedOrder   []string //合单ID
}

type OrderBuyValue struct {
	ClientOrderId       string  //客户端订单ID
	OfficialOrderId     int64   //交易所订单ID
	BuySize             float64 //下单设置的买入数量
	BuyType             string  //买入方式
	BuyPrice            float64 //下单价格
	BuyTime             int64   //下单时间
	FilledSize          float64 //成交数量
	FilledPrice         float64 //成交价格
	LimitSize           float64 //挂单数量
	CanceledSize        int32   //撤单数量
	CanceledTime        int64   //挂单的撤单时间
	CoefficientStopLoss float64 //止损系数
	BuyOrderStatus      int8    //订单状态
}

type OrderSellValue struct {
	ClientOrderId   string  //客户端订单ID
	OfficialOrderId int64   //交易所订单ID
	SellSize        float64 //下单设置的卖出数量
	SellType        string  //卖出方式
	SellPrice       float64 //下单价格
	SellTime        int64   //下单时间
	FilledSize      float64 //成交数量
	FilledPrice     float64 //成交价格
	LimitSize       float64 //挂单数量
	CanceledSize    int32   //撤单数量
	CanceledTime    int64   //挂单的撤单时间
}

type OrderSell struct {
	clientOid	string
	volHold		float64
	price		float64
	orderType	string
	available	bool
}

type instrumentHoldPair struct {
	key		string
	value 	*InstrumentHoldValue
}

type instrumentHoldPairList []instrumentHoldPair

func(p instrumentHoldPairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p instrumentHoldPairList) Len() int {
	return len(p)
}

func (p instrumentHoldPairList) Less(i, j int) bool {
	if p[i].value.OrderStatus2 == p[j].value.OrderStatus2 {
		return -p[i].value.LineStopLoss < -p[j].value.LineStopLoss
	}
	return -p[i].value.OrderStatus2 < -p[j].value.OrderStatus2
}

func sortMapByValue(m map[string]*InstrumentHoldValue) instrumentHoldPairList {
	p := make(instrumentHoldPairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = instrumentHoldPair{k, v}
	}
	sort.Sort(p)
	return p
}

func cancel_order(string, int64) (int64, error){
	return 0, nil
}

func create_order(symbol string, accountId string,
	oderType string, source string, amount float64,
	price float64, clientOrderId string) (string, error) {
	return "", nil
}

func get_asks_bids_price(string, float64) (float64, float64, float64) {
	return 0, 0, 0
}