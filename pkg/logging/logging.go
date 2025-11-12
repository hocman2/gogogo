package logging

import (
  "runtime"
  "fmt"
)

func AddLocation(in string) string {
  pc, file, line, ok := runtime.Caller(1); 
  if !ok {
    return in;
  }

  fn := runtime.FuncForPC(pc);
  return fmt.Sprintf("%s:%d > %s: %s", file, line, fn.Name(), in);
}
