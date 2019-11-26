#!/bin/bash

#./_output/bin/swagger-gen \
    #--input-dirs yunion.io/x/onecloud/pkg/compute/models \
    #--input-dirs yunion.io/x/onecloud/pkg/image/models \
    #--output-package yunion.io/x/onecloud/pkg/generated/swagger/compute

./_output/bin/swagger-gen \
    --input-dirs yunion.io/x/onecloud/pkg/keystone/tokens \
    --input-dirs yunion.io/x/onecloud/pkg/keystone/models \
    --output-package yunion.io/x/onecloud/pkg/generated/swagger/identity
