#include "duktape.h"
#include "_cgo_export.h"

duk_ret_t http_request(duk_context *ctx)
{
  return httpRequest(ctx);
}

duk_c_function http_req() {
  return http_request;
}

