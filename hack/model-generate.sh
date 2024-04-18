#!/bin/bash

./_output/bin/model-api-gen --input-dirs yunion.io/x/onecloud/pkg/cloudcommon/db --output-package yunion.io/x/onecloud/pkg/apis
./_output/bin/model-api-gen --input-dirs yunion.io/x/cloudmux/pkg/cloudprovider --output-package yunion.io/x/onecloud/pkg/apis/cloudprovider
./_output/bin/model-api-gen --input-dirs yunion.io/x/onecloud/pkg/compute/models --output-package yunion.io/x/onecloud/pkg/apis/compute
./_output/bin/model-api-gen --input-dirs yunion.io/x/onecloud/pkg/image/models --output-package yunion.io/x/onecloud/pkg/apis/image
