; ModuleID = 'main'
source_filename = "main"

%"github.com/goplus/llgo/internal/runtime.Slice" = type { ptr, i64, i64 }

@"main.init$guard" = global ptr null
@__llgo_argc = global ptr null
@__llgo_argv = global ptr null
@0 = private unnamed_addr constant [10 x i8] c"Hello %d\0A\00", align 1

define void @main.init() {
_llgo_0:
  %0 = load i1, ptr @"main.init$guard", align 1
  br i1 %0, label %_llgo_2, label %_llgo_1

_llgo_1:                                          ; preds = %_llgo_0
  store i1 true, ptr @"main.init$guard", align 1
  br label %_llgo_2

_llgo_2:                                          ; preds = %_llgo_1, %_llgo_0
  ret void
}

define void @main(i32 %0, ptr %1) {
_llgo_0:
  store i32 %0, ptr @__llgo_argc, align 4
  store ptr %1, ptr @__llgo_argv, align 8
  call void @"github.com/goplus/llgo/internal/runtime.init"()
  call void @main.init()
  %2 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocZ"(i64 32)
  %3 = getelementptr inbounds i64, ptr %2, i64 0
  store i64 1, ptr %3, align 4
  %4 = getelementptr inbounds i64, ptr %2, i64 1
  store i64 2, ptr %4, align 4
  %5 = getelementptr inbounds i64, ptr %2, i64 2
  store i64 3, ptr %5, align 4
  %6 = getelementptr inbounds i64, ptr %2, i64 3
  store i64 4, ptr %6, align 4
  %7 = call %"github.com/goplus/llgo/internal/runtime.Slice" @"github.com/goplus/llgo/internal/runtime.NewSlice3"(ptr %2, i64 8, i64 4, i64 0, i64 4, i64 4)
  %8 = call i64 @main.sum(%"github.com/goplus/llgo/internal/runtime.Slice" %7)
  %9 = call i32 (ptr, ...) @printf(ptr @0, i64 %8)
  ret void
}

define i64 @main.sum(%"github.com/goplus/llgo/internal/runtime.Slice" %0) {
_llgo_0:
  %1 = call i64 @"github.com/goplus/llgo/internal/runtime.SliceLen"(%"github.com/goplus/llgo/internal/runtime.Slice" %0)
  br label %_llgo_1

_llgo_1:                                          ; preds = %_llgo_2, %_llgo_0
  %2 = phi i64 [ 0, %_llgo_0 ], [ %9, %_llgo_2 ]
  %3 = phi i64 [ -1, %_llgo_0 ], [ %4, %_llgo_2 ]
  %4 = add i64 %3, 1
  %5 = icmp slt i64 %4, %1
  br i1 %5, label %_llgo_2, label %_llgo_3

_llgo_2:                                          ; preds = %_llgo_1
  %6 = call ptr @"github.com/goplus/llgo/internal/runtime.SliceData"(%"github.com/goplus/llgo/internal/runtime.Slice" %0)
  %7 = getelementptr inbounds i64, ptr %6, i64 %4
  %8 = load i64, ptr %7, align 4
  %9 = add i64 %2, %8
  br label %_llgo_1

_llgo_3:                                          ; preds = %_llgo_1
  ret i64 %2
}

declare void @"github.com/goplus/llgo/internal/runtime.init"()

declare ptr @"github.com/goplus/llgo/internal/runtime.AllocZ"(i64)

declare %"github.com/goplus/llgo/internal/runtime.Slice" @"github.com/goplus/llgo/internal/runtime.NewSlice3"(ptr, i64, i64, i64, i64, i64)

declare i32 @printf(ptr, ...)

declare i64 @"github.com/goplus/llgo/internal/runtime.SliceLen"(%"github.com/goplus/llgo/internal/runtime.Slice")

declare ptr @"github.com/goplus/llgo/internal/runtime.SliceData"(%"github.com/goplus/llgo/internal/runtime.Slice")
