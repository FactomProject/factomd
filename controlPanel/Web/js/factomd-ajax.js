function queryState(item, value, func) {
  var req = new XMLHttpRequest()

  req.onreadystatechange = function() {
    if(req.readyState == 4) {
      //console.log(item + " - " + req.response)
      func(req.response)
    }
  }
  req.open("GET", "/factomd?item=" + item + "&value=" + value, true)
  req.open("GET", "/factomd?item=" + item + "&value=" + value, true)
  req.send()
}

$("#factom-search").click(function() {
  $(".factom-search-error").slideUp( 300 )
})

$("#factom-search-submit").click(function() {
  var x = new XMLHttpRequest()
  x.onreadystatechange = function() {
    if(x.readyState == 4) {
      obj = JSON.parse(x.response)
      if (obj.Type != "None") {
        redirect("search?input=" + $("#factom-search").val() + "&type=" + obj.Type, "post", x.response) // Something found
      } else {
        $(".factom-search-error").slideDown(300)
        console.log(x.response)
      }
    }
  }
  var formData = new FormData();
  formData.append("method", "search")
  formData.append("search", $("#factom-search").val())

  x.open("POST", "/post")
  x.send(formData)
})

$("section #factom-search-link").on('click',function(e) {
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
      } else {
        $(".factom-search-error").slideDown(300)
        console.log(x.response)
      }
    }
  }
  var formDataLink = new FormData();
  formDataLink.append("method", "search")
  formDataLink.append("search", hash)

  x.open("POST", "/post")
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

// Example Code to use for forms
/*
var form = document.getElementById("test_form")
form.addEventListener("submit", function(e) {
  e.preventDefault()
  var x = new XMLHttpRequest()

  x.onreadystatechange = function() {
    if(x.readyState == 4) {
      //console.log(x.response)
      //alert(x.response)
      $("#changeme").text(x.response)
    }
  }

  x.open("POST", "/post")
  x.send(new FormData(form))
})
*/
