"use strict";

function dbg(x) {
  console.log(x);
  return x;
}

class AppState {
  constructor() {
    this.color = undefined;
    this.inTurn = false;
    this.selected_cell = null;
    this.socket = this.start_socket();
  }

  start_socket() {
    let socket = new WebSocket("ws://localhost:9990/game");

    socket.onopen = function (_e) {
      document.getElementById("notification").innerHTML =
        "Looking for opponent...";
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log(data);
      switch (data.message) {
        case "gameStart": {
          this.startGame(data.startFirst ? "white" : "black");
          document.getElementById("notification").innerHTML = "";
          break;
        }
        case "playerTurn": {
          const [from, to] = data.opponentMove;
          this.move(this.coordFromNotation(from), this.coordFromNotation(to));
          break;
        }

        case "moveAccepted":
          this.selected_cell = null;
          this.selected_target = null;
          this.move(this.selected_cell, this.selected_target);
          break;

        case "illegalMove": {
          this.selected_cell = null;
          this.selected_target = null;
          const notification = document.getElementById("notification");
          const notice = document.createElement("p", { class: "error" });
          notice.innerHTML = "Illegal Move";
          notification.replaceChildren(notice);
          break;
        }

        default:
          console.log("received unknown websocket data ", data);
          break;
      }
    };

    socket.onclose = function (event) {
      if (event.wasClean) {
        alert(
          `[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`,
        );
      } else {
        // e.g. server process killed or network down
        // event.code is usually 1006 in this case
        alert("[close] Connection died");
      }
    };

    socket.onerror = function (_error) {
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
    let home_row = [
      "rook",
      "knight",
      "bishop",
      "queen",
      "king",
      "bishop",
      "knight",
      "rook",
    ];
    if (color == "black") {
      home_row = home_row.reverse();
    }

    for (let i in home_row) {
      const type = home_row[i];
      row_0.children[i].appendChild(AppState.newUnit(opp_color, type));
      row_7.children[i].appendChild(AppState.newUnit(color, type));
    }

    for (const cell of document.getElementById("row-1").children) {
      cell.appendChild(AppState.newUnit(opp_color, "pawn"));
    }

    for (const cell of document.getElementById("row-6").children) {
      cell.appendChild(AppState.newUnit(color, "pawn"));
    }

    const board = document.getElementById("board-body");
    for (let i = 0; i < board.children.length; i++) {
      const row = board.children[i];
      for (let j = 0; j < row.children.length; j++) {
        const cell = row.children[j];
        cell.onclick = () => this.select_cell(j, i);
      }
    }
  }

  static newUnit(color, type) {
    const unit = document.createElement("div");
    unit.classList.add(color, type);
    return unit;
  }

  move(from, to) {
    const cfrom = document.getElementById(`cell-${from.x}-${from.y}`);
    console.assert(cfrom);
    const cto = document.getElementById(`cell-${to.x}-${to.y}`);
    console.assert(cto);
    const unit = cfrom.removeChild(cfrom.firstChild);
    cto.replaceChildren(unit);
  }

  select_cell(x, y) {
    console.log("selecting ", x, y);
    if (this.selected_cell == null) {
      const selection = document.getElementById(`cell-${x}-${y}`);
      console.assert(selection);
      if (
        selection.firstChild == null ||
        !selection.firstChild.classList.contains(this.color)
      ) {
        return;
      }
      this.selected_cell = { x: x, y: y };
    } else {
      this.selected_target = { x: x, y: y };
      this.socket.send(
        JSON.stringify({
          request: "move",
          move: [
            dbg(
              this.coordToNotation(this.selected_cell.x, this.selected_cell.y),
            ),
            dbg(this.coordToNotation(x, y)),
          ],
        }),
      );
    }
  }

  coordToNotation(x, y) {
    if (this.color == "white") {
      x = 7 - x;
      y = 7 - y;
    }

    return `${String.fromCodePoint("a".codePointAt(0) + x)}${y}`;
  }

  coordFromNotation(s) {
    let x = s.codePointAt(0) - "a".codePointAt(0);
    let y = s.codePointAt(1) - "1".codePointAt(0);

    if (this.color == "white") {
      x = 7 - x;
      y = 7 - y;
    }

    return { x: x, y: y };
  }
}

window.onload = function () {
  const tbody = document.getElementById("board-body");

  for (let i = 0; i < 8; i++) {
    const tr = document.createElement("tr");
    tr.id = `row-${i}`;

    for (let j = 0; j < 8; j++) {
      const td = document.createElement("td");
      td.id = `cell-${j}-${i}`;
      td.classList.add((i + j) % 2 == 0 ? "tile-dark" : "tile-light");
      tr.appendChild(td);
    }
    tbody.appendChild(tr);
  }

  new AppState();
  // const socket = start_socket();
};
