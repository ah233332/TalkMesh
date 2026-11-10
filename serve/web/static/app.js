let me = { userId: 0, userName: "" };

const list = document.getElementById("messages");

async function refresh() {
  try {
    const resp = await fetch("/api/messages", { credentials: "same-origin" });
    if (!resp.ok) {
      return;
    }
    const data = await resp.json();
    me = data.me || me;
    render(data.messages || []);
  } catch (e) {
    // 网络抖动时忽略，下一轮再试
  }
}

function render(messages) {
  const nearBottom =
    list.scrollHeight - list.scrollTop - list.clientHeight < 40;

  list.innerHTML = "";

  for (const m of messages) {
    const li = document.createElement("li");

    if (m.type === "group" || m.type === "private") {
      const mine = m.userId === me.userId;
      li.className = "msg " + (mine ? "me" : "other");

      const who = document.createElement("div");
      who.className = "who";
      who.textContent =
        (m.userName || "未知") +
        "(" + m.userId + ")" +
        (m.type === "private" ? " 私聊" : "");

      const bubble = document.createElement("div");
      bubble.className = "bubble";
      bubble.textContent = m.content || "";

      li.appendChild(who);
      li.appendChild(bubble);
    } else {
      li.className = "sys";
      li.textContent =
        (m.userName ? m.userName + "(" + m.userId + ") " : "") +
        (m.text || "");
    }

    list.appendChild(li);
  }

  if (nearBottom) {
    list.scrollTop = list.scrollHeight;
  }
}

refresh();
setInterval(refresh, 2000);
