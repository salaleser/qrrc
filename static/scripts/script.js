function answer() {
    document.location.href = "/spotify/nonamegamego/answer"
}

function answer0() {
    document.location.href = "/spotify/nonamegamego/main?correct=-50"
}

function answer1() {
    document.location.href = "/spotify/nonamegamego/main?correct=50"
}

function answer2() {
    document.location.href = "/spotify/nonamegamego/main?correct=100"
}

function exit() {
    document.location.href = "/spotify/home"
}

function start() {
    const names = []
    const c = document.getElementsByClassName('name')
    for (let i = 0; i < c.length; i++) {
        const v = c.item(i).value
        if (v !== '') {
            names.push(v)
        }
    }

    if (names.length < 2) {
        alert('Укажите более одного игрока')
        return
    }

    document.location.href = "/spotify/nonamegamego/main?start=true&player_names=" + names.join(',')
}   

function setPlaylist() {
    const v = document.getElementById('playlist').value
    if (v === '-') {
        v = ''
    }
    document.location.href = `/spotify/nonamegamego/setup?playlist=${v}`
}