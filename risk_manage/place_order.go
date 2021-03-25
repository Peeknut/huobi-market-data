package risk_manage

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const symbol_test = "btcusdt"

var (
	parameters_place_order[10]float64 = [10]float64{0.001,0.0002,0.9993}
	//parameters_place_order[0] = 0.001   # 交易所限定最小下单instrument量, minimum of orders when placing (ex. BTC)
	//parameters_place_order[1] = 0.0002  # 每单资金占比, max capitals of each order to total capitals
	//parameters_place_order[2] = 0.9993  # 止损系数, 计算止损线=买入价*止损价, coefficient_stopLoss
)

func placeOrder()  {
	for true {
		//####################################################
		//###### Place orders: including buy and sell orders
		//### Parameter of buy orders
		// 每单最大金额, max purchase amount
		//max_risks_transaction := capitals_account[0] * parameters_place_order[0]
		price_order := 0.0       // 下单价格, price to place the order
		amount_order := 0.0      // 下单的量, amount to place the order
		//type_order := "buy-ioc"  // 下单的类型, type to place the order, default is "buy_ioc"

		// ###### Place a buy order
		if signals_risk_action[0] == 0 && accountHold[1][3] >= parameters_place_order[0] * trade_price {
			//# 当前模型预测结果（0不涨1涨）
			if model_pred_data && is_pred_data_change[1] != 0 {
				is_pred_data_change[1] = 0
				random_amount := float64(GenerateRangeNum(0,1e6) * 1.0) / 1e6    // 小数点后保留6位
				random_amount =  0.001 * (1 + random_amount)  // 后续升级为动态调整下单量！！！
				amount_order, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", random_amount), 64)
				price_order, _, _ = get_asks_bids_price("buy", amount_order)
				if amount_order * price_order <= accountHold[1][3] {
					client_order_id := strconv.FormatInt(time.Now().UnixNano()/1e3,10)[3:16] + "b" + string(int(random_amount * 10000))
					temp_buy_order := &OrderBuyValue {
						ClientOrderId:			client_order_id,
						CanceledSize:			1,
						CoefficientStopLoss:	1 - target_para_PLR[1],
					}
					orderBuy[client_order_id[:13]] = temp_buy_order
					fmt.Println("[place_order] buy order message:\n", "OrderType.BUY_LIMIT: ", BUY_LIMIT,
						" amount_order: ", amount_order, " price_order: ", price_order, " client_order_id: ", client_order_id)
					//order_id, err := trade_client.create_order(symbol=symbol_test, account_id=account_id, order_type=BUY_LIMIT,
					//	source=OrderSource.API, amount=amount_order, price=price_order,
					//	client_order_id=client_order_id)
					order_id, err := create_order(symbol_test, account_id, BUY_LIMIT, OrderSource.API,
						amount_order, price_order, client_order_id)
					if err != nil {
						fmt.Println(err)
						delete(orderBuy, client_order_id[:13])
					} else {
						//LogInfo.output("created order id : {id}".format(id=order_id))
						fmt.Println("created order id : {id}", order_id)
					}
				}
			} else {
				is_pred_data_change[1] = 0
			}
		}

		//###### Place a sell order
		//### Check the list 'orders_sell' to sell orders
		for client_order_id, client_ids := range orders_sell {
			if !client_ids.available {
				fmt.Println("[place_order] sell order...", "client_order_id: ", client_order_id,
					"str(client_ids[0])",string(client_ids.clientOid),
					" amount: ", (int(client_ids.volHold * 100000)) * 1.0 / 100000, " price: ", &client_ids.price)
				client_ids.available = true
				temp_sell_order := &OrderSellValue{
					ClientOrderId: client_order_id,
				}
				orderSell[client_order_id] = temp_sell_order
				temp_amount := float64(int(client_ids.volHold * 100000)) * 1.0 / 100000
				//order_id, err := trade_client.create_order(symbol=symbol_test, account_id=account_id, order_type=OrderType.SELL_IOC,
				//	source=OrderSource.API, amount=temp_amount, price=client_ids.price,
				//	client_order_id=client_ids.clientOid)
				order_id, err := create_order(symbol_test, account_id, SELL_IOC, OrderSource.API,
					temp_amount, client_ids.price, client_ids.clientOid)
				if err != nil {
					fmt.Println(err)
					delete(orderSell, client_order_id)
				} else {
					//LogInfo.output("created order id : {id}".format(id=order_id))
					fmt.Println("created order id : {id}", order_id)
				}
			}
		}

		for client_order_id, order_cancel_id := range orders_cancel_buy {
			fmt.Println("orders_cancel_buy",orders_cancel_buy)
			//canceled_order_id, err := trade_client.cancel_order(symbol_test, order_cancel_id)
			canceled_order_id, err := cancel_order(symbol_test, order_cancel_id)
			if err != nil {
				fmt.Println("order cancel error...", err)
				delete(orders_cancel_buy, client_order_id)
				orderBuy[client_order_id].BuyOrderStatus = 0
			}
			if canceled_order_id == order_cancel_id {
				delete(orders_cancel_buy, client_order_id)
				//LogInfo.output("cancel order {id} done".format(id=canceled_order_id))
				fmt.Println("cancel order {id} done", canceled_order_id)
			} else {
				//LogInfo.output("cancel order {id} fail".format(id=canceled_order_id))
				fmt.Println("cancel order {id} fail", canceled_order_id)
			}
		}
	}
}

func GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max - min) + min
	return randNum
}
