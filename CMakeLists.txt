cmake_minimum_required(VERSION 3.18)
project(cudackpt LANGUAGES C CXX CUDA)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CUDA_STANDARD 17)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

find_package(CUDAToolkit REQUIRED)

add_library(cudackpt SHARED
  shim/interpose.c
  shim/tracker.cpp
  shim/ipc.c
  shim/snapshot.cu
  shim/restore.cu
  shim/module_io.cpp
  shim/coalesce.cpp
  shim/quiesce.cpp
  shim/pinned_pool.cpp
  shim/errcode.c
)

set_source_files_properties(shim/interpose.c PROPERTIES LANGUAGE CXX)
set_source_files_properties(shim/ipc.c PROPERTIES LANGUAGE CXX)

target_include_directories(cudackpt PRIVATE ${CMAKE_CURRENT_SOURCE_DIR}/shim)
target_compile_definitions(cudackpt PRIVATE _GNU_SOURCE)
target_link_libraries(cudackpt PRIVATE CUDA::cuda_driver dl pthread)

set_target_properties(cudackpt PROPERTIES
  OUTPUT_NAME cudackpt
  LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}
)

add_executable(vectoradd tests/integration/vectoradd.cu shim/app_hooks.c)
target_include_directories(vectoradd PRIVATE ${CMAKE_CURRENT_SOURCE_DIR}/shim)
target_link_libraries(vectoradd PRIVATE CUDA::cudart)

set_target_properties(vectoradd PROPERTIES
  RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}
)

add_executable(cublas_gemm tests/integration/cublas_gemm.cu shim/app_hooks.c)
target_include_directories(cublas_gemm PRIVATE ${CMAKE_CURRENT_SOURCE_DIR}/shim)
target_link_libraries(cublas_gemm PRIVATE CUDA::cudart CUDA::cublas)
set_target_properties(cublas_gemm PROPERTIES
  RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}
)

enable_testing()
add_executable(tracker_test tests/unit/tracker_test.cpp shim/tracker.cpp)
target_include_directories(tracker_test PRIVATE ${CMAKE_CURRENT_SOURCE_DIR}/shim)
target_link_libraries(tracker_test PRIVATE CUDA::cuda_driver)
add_test(NAME tracker COMMAND tracker_test)
