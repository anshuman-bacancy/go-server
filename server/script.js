document.getElementById("send").addEventListener("click", handleData, true);

function handleData() {
  console.log("reached the console")

  const name = document.getElementById("id")
  const comp = document.getElementById("comp")
  const pos = document.getElementById("pos")

  const http = new XMLHttpRequest();
  const url = "http://localhost:8000/"
  http.open("POST", url, true);
  http.setRequestHeader("Content-Type", "application/json");
  http.onreadystatechange = function() {
    if (http.readystate == 4 && http.status == 200) {
      console.log("data send")
    }
  }
  data = {
    "name": name,
    "comp": comp,
    "position": pos
  }
  http.send(data);
}
