//
// Function  : show_confirm_2
//
// Purpose   : Prompt the user for verification after they have clicked a submit button.
//
// Inputs    : the_form - the input form
//
// Output    : Confirmartion prompt
//
// Returns   : IF confirmation denied THEN false
//
// Example   : onclick="return show_confirm_2(document.myform1,'delete this item');"
//
// Notes     : (none)
//

function show_confirm_2(the_form,a_message)
{
var user_reply = confirm("Are you sure you want to " + a_message + " ?");
if ( user_reply == true ) {
  the_form.submit();
}
else {
  return false;
}
} // end of show_confirm_2
