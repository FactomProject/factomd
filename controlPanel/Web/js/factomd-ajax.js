function queryState(item, func) {
  var req = new XMLHttpRequest()

  req.onreadystatechange = function() {
    if(req.readyState == 4) {
      //console.log(item + " - " + req.response)
      func(req.response)
    }
  }
  req.open("GET", "/factomd?item=" + item, true)
  req.send()
}


$("#factom-search-submit").click(function() {
  var x = new XMLHttpRequest()
  x.onreadystatechange = function() {
    if(x.readyState == 4) {
      obj = JSON.parse(x.response)
      if (obj.Type != "None") {
        redirect("search?input=" + $("#factom-search").val(), "post", x.response) // Something found
      } else {
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

$("#factom-search-link").click(function() {
  type = jQuery(this).attr("type")
  hash = jQuery(this).text()
  alert(hash)
  var x = new XMLHttpRequest()
  x.onreadystatechange = function() {
    if(x.readyState == 4) {
      console.log(x.response)
      obj = JSON.parse(x.response)
      if (obj.Type != "None") {
        redirect("search?input=" + hash +"&type=" + type, "post", x.response) // Something found
      } else {
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
