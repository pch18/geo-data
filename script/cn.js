// https://github.com/adyliu/china_area
// id, parent, name, pinyin, post, address

//const py = require(`pinyin`);
const raw = require(`fs`).readFileSync(`./cn-raw.csv`, `utf8`)
x = raw.split(`\n`).map(l => l.split`,`).map(a => ({
  id: a[0],
  name: a[1],
  parent: a[3],
  pinyin:a[4]
})).filter(a => a.id && a.id.length === 12)

map = Object.fromEntries(x.map(a => [a.id, a]))


res = x.map(a => {
  let currentParent = a.parent

  // 处理不存在的父级
  while (currentParent && (!map[currentParent])) {
    currentParent = getParentId(currentParent)
  }

  let address = a.name
  let parent = map[currentParent]
  while (parent) {
    address = parent.name + address
    parent = map[parent.parent]
  }

 // const pinyin = py.pinyin(a.name, { style: "normal" }).flat().join(` `)
  return {
    id: a.id,
    parent: currentParent,
    name: a.name,
    pinyin: a.pinyin?.replace('\r',''),
    address
  }
})

const csv = res.map(a => `${a.id},${a.parent || ''},${a.name},${a.pinyin},${a.address || ''},`).join('\n')

require(`fs`).writeFileSync(`../data/cn.csv`, csv)

// 根据编码获取上级ID
function getParentId(id) {
  // 判断当前级别并获取上级
  if (id.endsWith('0000000000')) return null // 省级无上级
  if (id.endsWith('00000000')) return id.substring(0, 2) + '0000000000' // 地级->省级
  if (id.endsWith('000000')) return id.substring(0, 4) + '00000000' // 县级->地级
  if (id.endsWith('000')) return id.substring(0, 6) + '000000' // 乡级->县级
  return id.substring(0, 9) + '000' // 村级->乡级
}