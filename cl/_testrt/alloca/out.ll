; ModuleID = 'main'
source_filename = "main"

@"main.init$guard" = global ptr null
@0 = private unnamed_addr constant [4 x i8] c"Hi\0A\00", align 1
@1 = private unnamed_addr constant [3 x i8] c"%s\00", align 1

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

define void @main() {
_llgo_0:
  call void @main.init()
  %0 = alloca i8, i64 4, align 1
  %1 = call ptr @memcpy(ptr %0, ptr @0, i64 4)
  %2 = call i32 (ptr, ...) @printf(ptr @1, ptr %0)
  ret void
}

declare ptr @memcpy(ptr, ptr, i64)

declare i32 @printf(ptr, ...)