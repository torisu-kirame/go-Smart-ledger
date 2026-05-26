/** 入账/审批后从服务端拉取最新账本元数据与事件流水 */
export async function refreshLedgerFromServer(api, ledgerId) {
  const [ledger, events] = await Promise.all([
    api.getLedger(ledgerId),
    api.listEvents(ledgerId),
  ])
  return { ledger, events }
}
