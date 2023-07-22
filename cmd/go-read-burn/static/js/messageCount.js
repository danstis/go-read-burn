const text_max = 8000;
$("#input_count").html("Password or secret (0 / " + text_max + " chars):");

$("#inputText").keyup(function () {
  let text_length = $("#inputText").val().length;

  $("#input_count").html(
    "Password or secret (" + text_length + " / " + text_max + " chars):"
  );
});
