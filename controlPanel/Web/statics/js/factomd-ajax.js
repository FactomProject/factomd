function queryState(item, value, func) {
  var req = new XMLHttpRequest()

  req.onreadystatechange = function() {
    if(req.readyState == 4) {
      //console.log(item + " - " + req.response)
      func(req.response)
    }
  }
  req.open("GET", "./factomd?item=" + item + "&value=" + value, true)
  req.send()
}

function batchQueryState(item, func) {
  var req = new XMLHttpRequest()

  req.onreadystatechange = function() {
    if(req.readyState == 4) {
      //console.log(item + " - " + req.response)
      func(req.response)
    }
  }
  req.open("GET", "./factomdBatch?batch=" + item, true)
  req.send()
}

$("#factom-search").click(function() {
  $(".factom-search-error").slideUp( 300 )
})

$("#factom-search-submit").click(function() {
  searchBarSubmit()
})
$(".factom-search-container").keypress(function(e) {
  var key = e.which || e.keyCode;
  if (!(key == 13)) {
    return
  }
  searchBarSubmit()
})

function searchBarSubmit() {
  var x = new XMLHttpRequest()
  x.onreadystatechange = function() {
    if(x.readyState == 4) {
      console.log(x.response)
      obj = JSON.parse(x.response)
      if (obj.Type == "dblockHeight") {
        window.location = "search?input=" + obj.item + "&type=dblock"
      } else if (obj.Type != "None") {
        window.location = "search?input=" + $("#factom-search").val() + "&type=" + obj.Type
       //redirect("search?input=" + $("#factom-search").val() + "&type=" + obj.Type, "post", x.response) // Something found
      } else {
        $(".factom-search-error").slideDown(300)
        console.log(x.response)
      }
    }
  }
  var formData = new FormData();
  formData.append("method", "search")
  formData.append("search", $("#factom-search").val())

  x.open("POST", "./post")
  x.send(formData)
}

$("body").on('mouseup',"section #factom-search-link",function(e) {
  type = jQuery(this).attr("type")
  hash = jQuery(this).text()
  var x = new XMLHttpRequest()
  x.onreadystatechange = function() {
    if(x.readyState == 4) {
      obj = JSON.parse(x.response)
      if (obj.Type != "None") {
        if(e.which == 1){
          window.location = "search?input=" + hash + "&type=" + type
        } else if(e.which == 2) {
          window.open("/search?input=" + hash + "&type=" + type);
        }
        //redirect("search?input=" + hash + "&type=" + type, "post", x.response) // Something found
      } else if(obj.Type == "special-action-fack"){
        window.location = "search?input=" + hash + "&type=" + type
      } else {
        $(".factom-search-error").slideDown(300)
        //console.log(x.response)
      }
    }
  }
  var formDataLink = new FormData();
  formDataLink.append("method", "search")
  formDataLink.append("search", hash)
  formDataLink.append("known", type)

  x.open("POST", "./post")
  x.send(formDataLink)
})

// Redirect with post content
function redirect(url, method, content) {
  var input = $("<input>").attr("type", "hidden").val(content).attr("name", "content")
  var x = $('<form>', {
      method: method,
      action: url
  })
  x.append(input)
  x.submit();
};


function nextNode() {
  resp = queryState("nextNode","",function(resp){
    $("#current-node-number").text(resp)
  })
}
