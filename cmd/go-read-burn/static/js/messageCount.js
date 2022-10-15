var text_max = 8000;
$("#input_count").html("Password or secret (0 / " + text_max + " chars):");

$("#inputText").keyup(function () {
  var text_length = $("#inputText").val().length;
  var text_remaining = text_max - text_length;

  $("#input_count").html(
    "Password or secret (" + text_length + " / " + text_max + " chars):"
  );
});
