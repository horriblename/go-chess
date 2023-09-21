"use strict"


class AppState {
  constructor() {
    this.color = undefined;
    this.inTurn = false;
    this.selected_cell = null;
    this.socket = this.start_socket();
  }

  start_socket() {
    let socket = new WebSocket("ws://localhost:9990/game");

    socket.onopen = function(_e) { };

    socket.onmessage = function(event) {
      switch (event.data.message) {
        case "gameStart": {
          this.startGame(event.data.startFirst ? "white" : "black");
          break;
        }
        case "playerTurn": {
          const [from, to] = event.data.opponentmove;
          this.move(coordFromNotation(from), coordFromNotation(to));
          break;
        }

        case "moveAccepted":
          this.move(this.selected_cell, this.selected_target);
          break;

        case "illegalMove": {
          const notification = document.getElementById("notification");
          const notice = document.createElement("p", { "class": "error" });
          notice.innerHTML = "Illegal Move"
          notification.replaceChild(notification.firstChild, notice);
          break;
        }

        default:
          console.log("received unknown websocket data ", event.data)
          break;
      }
    };

    socket.onclose = function(event) {
      if (event.wasClean) {
        alert(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
      } else {
        // e.g. server process killed or network down
        // event.code is usually 1006 in this case
        alert('[close] Connection died');
      }
    };

    socket.onerror = function(_error) {
      alert(`[error]`);
    };
    return socket;
  }

  startGame(color) {
    if (color != "black" && color != "white") {
      console.trace("Error: invalid color");
      color = "white";
    }
    this.color = color;
    const opp_color = color == "white" ? "black" : "white";

    const row_0 = document.getElementById("row-0");
    const row_7 = document.getElementById("row-7");
    let home_row = ["castle", "knight", "bishop", "queen", "king", "bishop", "knight", "castle"];
    if (color == "black") {
      home_row = home_row.reverse();
    }

    for (let i in home_row) {
      const type = home_row[i];
      row_0.childNodes[i].replaceChildren(this.newUnit(opp_color, type));
      row_7.childNodes[i].replaceChildren(this.newUnit(color, type));
    }

    for (const cell of document.getElementById("row-1")) {
      cell.replaceChildren(this.newUnit(opp_color, "pawn"));
    }

    for (let cell of document.getElementById("row-6")) {
      cell.replaceChildren(this.newUnit(color, "pawn"));
    }
  }

  static newUnit(color, type) {
    const unit = document.createElement("span");
    unit.classList.add(color, type);
    return unit;
  }

  move(from, to) {
    const cfrom = document.getElementById(`cell-${from.x}-${from.y}`);
    const cto = document.getElementById(`cell-${to.x}-${to.y}`);
    const unit = cfrom.removeChild(cfrom.firstChild);
    cto.replaceChild(cto.firstChild, unit);

  }

  select_cell(x, y) {
    if (this.selected_cell == null) {
      const selection = document.getElementById(`cell-${x}-${y}`);
      if (selection.firstChild == null || !selection.firstChild.classList.contains(this.color)) {
        return;
      }
      this.selected_cell = { x: x, y: y };
    } else {
      this.selected_target = { x: x, y: y };
      this.socket.send({
        "request": "move", "move": [
          coordToNotation(this.selected_cell.x, this.selected_cell.y),
          coordToNotation(x, y)
        ]
      })
    }
  }
}

function coordToNotation(x, y) {
  if (this.color == 'white') {
    return `${String.fromCodePoint('a'.codePointAt(0) + x)}${y}`;
  } else {
    x = 7 - x;
    y = 7 - y;
    return `${String.fromCodePoint('a'.codePointAt(0) + x)}${y}`;
  }
}

function coordFromNotation(s) {
  let x = s.codePointAt(0) - 'a'.codePointAt(0);
  let y = s.codePointAt(1) - '1'.codePointAt(0);

  if (this.color == 'white') {
    x = 7 - x;
    y = 7 - y;
  }

  return { x: x, y: y };
}

window.onload = function() {
  const tbody = document.getElementById('board-body');

  for (let i = 0; i < 8; i++) {
    const tr = document.createElement('tr');
    tr.id = `row-${i}`;

    for (let j = 0; j < 8; j++) {
      const td = document.createElement('td');
      td.id = `cell-${j}-${i}`;
      td.classList.add((i + j) % 2 == 0 ? "tile-dark" : "tile-light")
      tr.appendChild(td)
    }
    tbody.appendChild(tr);
  }

  new AppState();
  // const socket = start_socket();
}
