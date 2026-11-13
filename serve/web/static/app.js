let me = { userId: 0, userName: "" };

const list = document.getElementById("messages");
const userList = document.getElementById("users");
const toUserIdInput = document.getElementById("toUserId");

async function refresh() {
  try {
    const resp = await fetch("/api/messages", { credentials: "same-origin" });
    if (!resp.ok) {
      return;
    }
    const data = await resp.json();
    me = data.me || me;
    render(data.messages || []);
    renderUsers(data.users || []);
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
    const timeEl = document.createElement("span");
    timeEl.className = "time";
    timeEl.textContent = m.time || "";

    if (m.type === "group" || m.type === "private") {
      const mine = m.userId === me.userId;
      li.className = "msg " + (mine ? "me" : "other");

      const who = document.createElement("div");
      who.className = "who";
      who.textContent =
        (m.userName || "未知") +
        "(" + m.userId + ")" +
        (m.type === "private" ? " 私聊" : "");

      who.appendChild(timeEl); 

      const bubble = document.createElement("div");
      bubble.className = "bubble";
      bubble.textContent = m.content || "";

      li.appendChild(who);
      li.appendChild(bubble);
    } else {
      li.className = "sys";
      li.textContent =
        (m.userName ? m.userName + "(" + m.userId + ") " : "") +
        (m.text || "")+" "+(m.time || "");
    }

    list.appendChild(li);
  }

  if (nearBottom) {
    list.scrollTop = list.scrollHeight;
  }
}


function renderUsers(users) {
  userList.innerHTML = "";

  for (const u of users) {
    const li = document.createElement("li");
    li.textContent = (u.userName || "未知") + "(" + u.userId + ")";
    li.title = "点击填入私聊对象";

    if (u.userId === me.userId) {
      li.classList.add("self");
    }

    li.addEventListener("click", function () {
      if (toUserIdInput) {
        toUserIdInput.value = u.userId;
      }
    });

    userList.appendChild(li);
  }
}

refresh();
setInterval(refresh, 2000);
