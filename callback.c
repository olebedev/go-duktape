#include "duktape.h"

duk_ret_t my_addtwo(duk_context *ctx) {
  double a, b;

  /* Here one can expect that duk_get_top(ctx) == 2, because nargs
   * for duk_push_c_function() is 2.
   */

  a = duk_get_number(ctx, 0);
  b = duk_get_number(ctx, 1);
  duk_push_number(ctx, a + b);
  return 1;   /*  1 = return value at top
               *  0 = return 'undefined'
               * <0 = throw error (use DUK_RET_xxx constants)
               */
}

duk_c_function go_duk_c_function() {
  return my_addtwo;
}
