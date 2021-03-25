package risk_manage

import (
	"fmt"
	"strconv"
	"time"
)

var (
	capitals_account [2]float64  // 数组:初始账户资金\当前账户资金
	orders_sell map[string]*OrderSell  // 买单
	orders_cancel_buy map[string]int64 // 卖单
	target_para_PLR [3]float64 = [...]float64{1.0023, 0.9993, 0.0005}
	target_PLR = (target_para_PLR[0] * (1 - target_para_PLR[2]) - 1 - target_para_PLR[2]) / (1 + target_para_PLR[2] - target_para_PLR[1] * (1 - target_para_PLR[2]))
	accumulation_pred = 0

	//###### Parameters for risk management
	//# array[0] = 风险资本比例, ratio of venture to total capital
	//# array[1] = 单笔交易最大风险, maximal risks taken in a single transaction
	//# array[2] = 所有未平仓仓位的最大整体风险, maximum of the overall risks in all holding trade targets
	//# array[3] = 停止交易的亏损上限, ratio of maximal loss to total venture capital to stop trading
	//# array[4] = 当前资金占初始资金的比例，ratio of current capital to original capital
	//# array[5] = 资金增长的调整参数时机, time to alter the value of original capitals when capitals increases
	//# array[6] = 资金减少的调整参数时机, time to alter the value of original capitals when capitals decreases
	//# array[7] = 期望收益阈值, threshold to expected profits
	//# array[8] = 交易统计阈值, threshold for trading statistics
	//# array[9] = 权益动量管理系数, coefficient of Equity momentum management
	//# array[10] = 模型准确率阈值, threshold to prediction accuracy of the model
	//# array[11] = 盈亏比阈值，threshold to profit / loss ratio
	parameters_risk_mng [12]float64 = [...]float64{1, 0.0002, 0.05, 0.1, 0, 1.15, 0.95, 30, 1.2, 0.95, 0.52, 0.9}

	//###### Parameters for orders management
	//# array[0] = 盈利时止损线调整的偏移量, bias for modulating the stop loss line when having profits
	//# array[1] = 止损预警偏移量, bias for stopping loss when the current price is touching the stop loss line
	//# array[2] = 止盈和主动止损偏移量, bias for keeping profits and active stop loss
	parameters_orders_mng [5]float64 = [5]float64{0.0001, 0.2, 0.0002}
)




func risk_mng() {
	time.Sleep(time.Second * 5)
	fmt.Println("accountHold: ", accountHold)
	capitals_account[0] = accountHold[1][1] + accountHold[0][1] * trade_price
	fmt.Println("[risk_mng] time sleep over, get capitals_account[0]: ", capitals_account[0])
	max_equity_momentum := 0.0 // 初始化权益动量, initial equity momentum
	for true {
		// ###### Capitals
		// # array[0] = 初始资金, original capitals
		// # array[1] = 当前资金, current capitals
		capitals_account[1] = accountHold[1][1] + accountHold[0][1] * trade_price

		// ###### Parameters for risk management
		if capitals_account[0] == 0 {
			fmt.Println("the initial capitals_account is 0")
			return
		}
		parameters_risk_mng[4] = capitals_account[1] / capitals_account[0]

		//step2：更新信号
		// ###### Signals for risk management
		// signals_risk_mng
		// signals_risk_action

		// ###### Signals for stop buying
		// ### array[0,0] = Signal 1# 风险资本金是否超限；array[0,1] = Signal 2# 所有未平仓仓位风险是否超限；
		// ### array[0,2] = Signal 3# 期望收益管理；array[0,3] = Signal 4# 权益动量管理
		// ### array[0,4] = Signal 5# 模型跟踪结果；array[0,5] = Signal 6# 盈亏比 <= 阈值
		// # 1# signal: if 现金余额 < 初始资金 * (1 - 风险资本比例), signal = 1
		// #            if account balance < origin capitals * (1 - ratio of venture to total capital)
		if accountHold[1][1] < capitals_account[0] * (1 - parameters_risk_mng[0]) {
			signals_risk_mng[0] = []int{1,0,0,0,0,0}
			signals_risk_action[0] = 1
		}
		// # 2# signal: if 所有未平仓仓位的最大整体风险 > 初始资金 * 未平仓仓位的风险阈值，signal = 1
		// #            if the risks of the overall holding targets >= the threshold
		// # 当前账户资金=USDT+BTC*price, equivalent amount of current account in USDT
		// # hold_usdt = capitals_account[1]
		// # 当前仓位整体风险, overall current risks in all holding trade targets
		// # 所有订单承担风险的总和 = sum( (costHold - stopLoss) * vol_hold)
		// #                       = sum((instrumentHold[:,3] - instrumentHold[:,5]) .* instrumentHold[:,6])
		amount_risk_holdings := 0.0
		for _, value := range instrumentHold {
			amount_risk_holdings += (value.CostHold - value.LineStopLoss) * value.VolStopLoss
		}
		if amount_risk_holdings >= capitals_account[0] * parameters_risk_mng[2] {
			signals_risk_mng[0] = []int{0,1,0,0,0,0}
			signals_risk_action[0] = 1
		}
		if (profit[0][1] + profit[1][1]) >= parameters_risk_mng[7] {
			// # 3# signal: if 期望收益值 <= 阈值
			// #            if expected profits <= threshold
			// # profit = [[总收益额][盈利单数],[总亏损额][亏损单数]]
			expected_profits := profit[0][0] / profit[1][0]
			if expected_profits <= parameters_risk_mng[8] {
				signals_risk_mng[0] = []int{0,0,1,0,0,0}
				signals_risk_action[0] = 1
			}
			// # 4# signal: Equity momentum management
			// #            if 当前期望收益 <= 最高期望收益 * 系数
			// #            if current expected profits <= max expected profits * coefficient
			equity_momentum := profit[0][0] - profit[1][0]
			if equity_momentum > max_equity_momentum {
				max_equity_momentum = equity_momentum
			}
			if equity_momentum <= max_equity_momentum * parameters_risk_mng[9] {
				signals_risk_mng[0] = []int{0,0,0,1,0,0}
				signals_risk_action[0] = 1
			}
			// # 5# signal: if 模型正确率 <= 阈值
			// # 5# signal: if current accuracy <= threshold
			var curr_acc float64
			if confusion_list[1][0] + confusion_list[1][1] != 0 {
				curr_acc = float64(confusion_list[1][1]) / float64((confusion_list[1][0]) + confusion_list[1][1])
			}
			if curr_acc <= parameters_risk_mng[10] {
				signals_risk_mng[0] = []int{0,0,0,0,1,0}
				signals_risk_action[0] = 1
			}
			// # 这种写法只有因为模型准确率停止风控的才能重启
			if signals_risk_mng[0][4] == 1 && curr_acc > parameters_risk_mng[10] {
				signals_risk_mng[0] = []int{0,0,0,0,0,0}
				signals_risk_action[0] = 0
			}
			//# 6# signal: if 盈亏比 <= 阈值
			//# 6# signal: if current profit / loss ratio <= threshold * target profit / loss ratio
			//### Profit / loss ratio
			//# 盈亏比 = (现价 + 费用 - 止损价 + 费用) / (目标价 - 费用 - 现价 - 费用)
			//# profit / loss ratio = (current price * (1- transaction cost)）- (current price * (1 - stop loss line) * (1 + transaction cost)))
			//#                       / [(target price * (1 - transaction cost)) - (current price * (1 + transaction cost))
			//# ex. target price = 1.0023 * current price, stop loss line = 0.9993 * current price, transaction cost = 0.0002
			//# num_para_PLR = 3
			//# num_models = 1
			//### Target profit / loss ratio
			//# target_para_PLR = np.zeros((num_models, num_para_PLR), dtype=np.float32)
			//# target_para_PLR[0, :] = [1.0023, 0.9993, 0.0002]
			//target_PLR := (target_para_PLR[0] * (1 - target_para_PLR[2]) - 1 - target_para_PLR[2]) / (1 + target_para_PLR[2] - target_para_PLR[1] * (1 - target_para_PLR[2]))
			// ### Current profit / loss ratio
			// # 当前模型的盈亏比，平均盈利 / 平均亏损 = (总收益额 / 盈利单数) / (总亏损额 / 亏损单数)
			// # profit / loss ratio correspondence to a certain model equals average profits divided by average loss
			// # overall profits / number of profit orders divided by overall loss / number of loss orders
			var current_PLR float64
			if profit[0][1] == 0 {
				current_PLR = 0
			} else if profit[1][1] == 0 {
				current_PLR = 2 * target_PLR
			} else {
				current_PLR = profit[0][0] * profit[1][1] / profit[0][1] / profit[1][0]
			}
			if current_PLR <= target_PLR * parameters_risk_mng[11] {
				signals_risk_mng[0] = []int{0,0,0,0,0,1}
				signals_risk_action[0] = 1
			}
		}

		// ###### Signals for stop trading
		// ### array[1,0] = Signal 1# 亏损金额超过停止交易的亏损上限；
		// ### 1# signal: if (1 - 当前盈亏额) > 停止交易的亏损上限
		// ### 1# signal: if (1 - current profit and loss ratio) > ratio of maximal loss to total venture capital
		if (1 - parameters_risk_mng[4]) >= parameters_risk_mng[3] {
			signals_risk_mng[1] = []int{1,0,0,0,0,0}
			signals_risk_action[1] = 1
		}

		// ###### Signals for parameters modulation
		// ### array[2,0] = Signal 1# 资金增长20%; array[2,1] = Signal 2# 资金减少10%
		// # 1# signal: if 当前盈利额 >= 调整参数上限
		// # 1# signal: if current profit >= (1 + up bound of parameter alteration when capitals increases)
		if parameters_risk_mng[4] >= parameters_risk_mng[5] {
			signals_risk_mng[2] = []int{1,0,0,0,0,0}
			signals_risk_action[2] = 1
		}
		// # 2# signal: if 当前亏损额 <= 调整参数下限
		// # 2# signal: if current loss <= (1 + lower limit of parameter alteration when capitals increases)
		if parameters_risk_mng[4] <= parameters_risk_mng[6] {
			signals_risk_mng[2] = []int{0,1,0,0,0,0}
			signals_risk_action[2] = 1
		}

		//step3：修改订单
		//########################################################
		// ###### Manage orders and their stop loss line
		// # 卖单格式, [client order id, amount_selling, price_selling, type_selling]
		// #          [客户id，卖币数量, 卖价，卖单类型]
		// # orders_sell = []     # 止损/止盈卖单, orders touching the stop loss line or keeping profits to be sold
		var price_fall float64 // price as an anchor to check whether the price is really falling or not
		//temp_orders_sell_stoploss := []  // temporary list to collect parameters of a passive stop loos selling order
		//temp_orders_sell_keeprofit := [] // temporary list to collect parameters of a active keep profit selling order
		// ###### Manage buy orders
		for _, value := range orderBuy {
			now_time := time.Now().Unix()  // 单位是秒
			if now_time > value.CanceledTime && value.BuyOrderStatus == 0 && value.CanceledTime != 0 {  // 【12】订单是可操作的
				orders_cancel_buy[value.ClientOrderId] = value.OfficialOrderId
				value.BuyOrderStatus = 1
				// 下面做的跟上面的操作是一样的结果，如果 orderBuy 用的是 slice 的话
				//temp_order_buy_line := orderBuy[client_order_id]
				//temp_order_buy_line["main"][12] = 1
				//orderBuy[client_order_id] = temp_order_buy_line
			}
		}
		//###### Manage sell orders
		//# Modulate stop line dynamically
		//# Scenario 1:   保本线
		//#               现价 >= (1 + 买卖综合交易费用 + 偏移量) * 买入价(成本)，止损线调整至(1 + 买卖综合交易费用) * 买入价(成本)
		//#               current price >= (1 + 0.0004 + 0.0001) * current price, stop loss line = (1 + 0.0004) * current price
		//# Scenario 2:   持盈线
		//#               现价 >= (1 + 目标收益率 + 偏移量) * 买入价(成本)，止损线调整至(1 + 目标收益率) * 买入价(成本)
		//#               current price >= (1.0023 + 0.0001) * current price + 0.5, stop loss line = 1.0023 * current price
		//# Scenario 3：  启动止损 Launch to sell the orders which touch the stop loss line
		//#               现价 <= (止损线 + 偏移量)，添加入下单列表
		//#               current price <= stop loss line + $ 0.2, append the correspondence orders to the list
		//# target_para_PLR[0, :] = [1.0023, 0.9993, 0.0002]
		//# print("[thread_risk_mng] Modulate stop line dynamically...")
		for instrument_id, line_orders := range instrumentHold {
			//line_orders := value.main
			if line_orders.OrderStatus == 0 {  // 订单为可下单状态
				if (line_orders.TypeStopLoss != 2) && (trade_price >= (target_para_PLR[0] + parameters_orders_mng[0])* line_orders.CostHold) {
					// 调整止损线到目标持盈线 modulate stop loss line to target profits
					line_orders.LineStopLoss = target_para_PLR[0] * line_orders.CostHold
					line_orders.TypeStopLoss = 2
					line_orders.OrderStatus2 = 1
				} else if (line_orders.TypeStopLoss != 1) && (trade_price >= (1 + 2 * target_para_PLR[2] + parameters_orders_mng[0]) * line_orders.CostHold) {
					// 调整止损线到保本线 modulate stop l[iters_hold]oss line to break-even value
					line_orders.LineStopLoss = (1 + 2 * target_para_PLR[2]) * line_orders.CostHold
					line_orders.TypeStopLoss = 1
					line_orders.OrderStatus2 = 1
				} else if trade_price <= line_orders.LineStopLoss + parameters_orders_mng[1]{
					// 现价触及止损线，启动止损信号 if current price touch the stop loss line, take actions to sell instruments
					line_orders.TypeStopLoss = 0
					line_orders.OrderStatus2 = 0
					line_orders.OrderStatus = 1
					fmt.Println("[thread_risk_mng] current price touch the stop loss line, take actions to sell instruments...")
					// 止损指令 [0] = client id [1] = amount [2] = price [3] = type [4] = available
					price, _, _ := get_asks_bids_price("sell", line_orders.VolHold)
					temp_orders_sell_stoploss := &OrderSell{
						clientOid: 		line_orders.ClientOrderId+"s"+strconv.FormatInt(time.Now().UnixNano()/1e3,10)[12:16],
						volHold:		line_orders.VolHold,
						price: 			price,
						orderType: 		SELL_IOC,
						available: 		false,
					}
					for remain_id, remain_order := range remainHold {
						//remain_order_main := remain_order.main
						if remain_order.OrderStatus == 0 {
							fmt.Println("[thread_risk_mng] combine remain order...")
							remain_order.OrderStatus = 1
							temp_orders_sell_stoploss.volHold += remain_order.VolHold
							temp_instrumentHold_line := instrumentHold[instrument_id].CombinedOrder
							if temp_instrumentHold_line == nil {
								temp_instrumentHold_line = make([]string, 0)
							}
							temp_instrumentHold_line = append(temp_instrumentHold_line, remain_id)
						}
					}
					orders_sell[temp_orders_sell_stoploss.clientOid] = temp_orders_sell_stoploss
				}
			}
		}
		//# Scenario 4：  启动止盈or主动止损 Launch to sell the orders focusing on keeping profits or active stopping loss
		//#               第一类情况，最大止损线为保本和止盈线, 模型第一次预测为跌(0)，记录当前价格为止跌价price_fall,
		//#                   如果当前价格 <= (1 - 偏移量) * 止跌价, 针对最大止损线的订单进行止盈；
		//#                   如果模型连续两次预测为跌(0)，对应两个最大止损线的订单进行止盈，依次类推；
		//#               section 1:  model prediction = price falling (0), evaluate current price to price_fall as an anchor
		//#                   if current price <= (1 - bias) * price_fall, sell the order with max stop loss line
		//#                   if two successive zero predictions, sell two orders with two largest stop loss line, and so on
		//#               第二类情况，最大止损线为初始止损线, 连续两次预测为跌(0)，并且当前价格 <= (1 - 2 * 偏移量) * 止跌价
		//#                   针对最大止损线的订单进行主动止损
		//#               section 2: current price <= price_fall * (1 - 2 * bias), append the correspondence orders to the list
		if !model_pred_data && is_pred_data_change[0] != 0 {  // Value from model_pred.py, denotes the results of predictions, 1 rise, 0 not rise
			is_pred_data_change[0] = 0
			accumulation_pred += 1
			if price_fall == 0 {
				price_fall = trade_price
			}
		} else {
			is_pred_data_change[0] = 0
			accumulation_pred = 0
			price_fall = 0.0
		}

		if accumulation_pred != 0 {
			current_price := trade_price
			if current_price <= (1 - parameters_orders_mng[2]) * price_fall {
				fmt.Println("[thread_risk_mng] section 1...")
				keeping_profits := sortMapByValue(instrumentHold)
				cnt_sell := 0
				for _, kp := range keeping_profits {
					if kp.value.TypeStopLoss == 0 {
						break
					}
					instrument_id := kp.key
					if instrumentHold[instrument_id].OrderStatus == 0 {
						fmt.Println("[thread_risk_mng] section 1: add orders_sell...")
						instrumentHold[instrument_id].OrderStatus = 1
						price, _, _ := get_asks_bids_price("sell", instrumentHold[instrument_id].VolHold)
						temp_orders_sell_keeprofit := &OrderSell{
							clientOid: 		instrumentHold[instrument_id].ClientOrderId+"s"+strconv.FormatInt(time.Now().UnixNano()/1e3,10)[12:16],
							volHold:		instrumentHold[instrument_id].VolHold,
							price: 			price,
							orderType: 		SELL_IOC,
							available: 		false,
						}
						for remain_id, remain_order := range remainHold {
							//remain_order_main := remain_order.main
							if remain_order.OrderStatus == 0 {
								fmt.Println("[thread_risk_mng] combine remain order...")
								remain_order.OrderStatus = 1
								temp_orders_sell_keeprofit.volHold += remain_order.VolHold
								temp_instrumentHold_line := instrumentHold[instrument_id].CombinedOrder
								if temp_instrumentHold_line == nil {
									temp_instrumentHold_line = make([]string, 0)
								}
								temp_instrumentHold_line = append(temp_instrumentHold_line, remain_id)
							}
						}
						orders_sell[temp_orders_sell_keeprofit.clientOid] = temp_orders_sell_keeprofit
						cnt_sell += 1
						if cnt_sell >= accumulation_pred {
							break
						}
					}
				}

			}
		}

		delete_ids := make([]string, 0)
		for orders_id, orders_line := range orders_sell {
			if orders_line.available {
				delete_ids = append(delete_ids, orders_id)
			}
		}
		for _, item := range delete_ids {
			delete(orders_sell, item)
		}
	}
}