// https://www.post.japanpost.jp/zipcode/dl/utf-zip.html
// id, parent, name, pinyin, post, address

const raw = require(`fs`).readFileSync(`./jp-raw.csv`, `utf8`)

x = raw.split(`\n`).map(l => l.split`,`).map(a => ({
    post: JSON.parse(a[2]),
    a1_yomi: JSON.parse(a[3]),
    a2_yomi: JSON.parse(a[4]),
    a3_yomi: JSON.parse(a[5]),
    a1: JSON.parse(a[6]),
    a2: JSON.parse(a[7]),
    a3: JSON.parse(a[8])
}))

i1 = 1
res1Map = {}
res1 = x.filter(a => {
    const name = a.a1
    if (!res1Map[name]) {
        res1Map[name] = i1++
        return true
    }
    return false
}).map(a => {
    const name = a.a1
    const id = res1Map[name]
    if (!id) {
        throw 'res1 id not found:' + name
    }

    return {
        id,
        level: 1,
        parent: 0,
        name: a.a1,
        yomi: a.a1_yomi,
        address: a.a1
    }
})

i2 = 1000
res2Map = {}
res2 = x.filter(a => {
    const name = a.a1 + '|' + a.a2
    if (!res2Map[name]) {
        res2Map[name] = i2++
        return true
    }
    return false
}).map(a => {
    const name = a.a1 + '|' + a.a2
    const id = res2Map[name]
    if (!id) {
        throw 'res2 id not found:' + name
    }

    const parentName = a.a1
    const parent = res1Map[parentName]
    if (!parent) {
        throw 'res2 parent not found:' + parentName
    }

    return {
        id,
        level: 2,
        parent,
        name: a.a2,
        yomi: a.a2_yomi,
        address: a.a1 + a.a2
    }
})

i3 = 100000
res3 = x.map(a => {
    let name = a.a3
    let yomi = a.a3_yomi
    let address = a.a1 + a.a2 + a.a3.replace(/（.*）/g, '')
    if (name === "以下に掲載がない場合") {
        yomi = "#"
        address = a.a1 + a.a2
    }

    const parentName = a.a1 + '|' + a.a2
    const parent = res2Map[parentName]
    if (!parent) {
        throw 'res3 parent not found:' + parentName
    }

    return {
        id: i3++,
        level: 3,
        parent,
        name,
        yomi,
        address,
        post: a.post,
    }
})

res = [...res1, ...res2, ...res3]
const csv = res.map(a => `${a.id},${a.parent || ''},${a.name},${a.yomi},${a.address || ''},${a.level},${a.post || ''}`).join('\n')

require(`fs`).writeFileSync(`../data/jp.csv`, csv)