package main

import (
	"context"
	"testing"
	"time"

	"github.com/awslabs/volume-modifier-for-k8s/pkg/controller"
	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/coordination/v1"
)

func TestSerialLeaseUpdate(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	
	mc := controller.NewMockModifyController(mockCtl)
	podName := "test-pod"
	ctx, cancel := context.WithCancel(context.Background())
	
	localName := "test-pod"
	fakeLocalLease := &v1.Lease{
		Spec: v1.LeaseSpec{
			HolderIdentity: &localName,
		},
	}
	fakeLocalLease.Name = "external-resizer-ebs-csi-aws-com"

	remoteName := "test2-pod"
	fakeRemoteLease := &v1.Lease{
		Spec: v1.LeaseSpec{
			HolderIdentity: &remoteName,
		},
	}
	fakeRemoteLease.Name = "external-resizer-ebs-csi-aws-com"

	// Should do nothing
	handleLeaseUpdate(fakeRemoteLease, podName, ctx, cancel, mc)	
	time.Sleep(1 * time.Second)

	// Should start controller
	mc.EXPECT().Run(gomock.Any(), gomock.Eq(ctx))
	handleLeaseUpdate(fakeLocalLease, podName, ctx, cancel, mc)	
	time.Sleep(1 * time.Second)

	// Should do nothing (controller is already started)
	handleLeaseUpdate(fakeLocalLease, podName, ctx, cancel, mc)	
	time.Sleep(1 * time.Second)

	// Should stop controller
	handleLeaseUpdate(fakeRemoteLease, podName, ctx, cancel, mc)	
	time.Sleep(1 * time.Second)

	// Context should have been cancelled
	select {
	case <-ctx.Done():
		// Do nothing, this means the context is correctly cancelled
	default:
		t.Fatalf("Context is NOT cancelled/finished when it should be")
	}

	// Should restart controller
	mc.EXPECT().Run(gomock.Any(), gomock.Eq(ctx))
	handleLeaseUpdate(fakeLocalLease, podName, ctx, cancel, mc)	
	time.Sleep(1 * time.Second)
	
	// Context that was just passed to restarted controller should NOT be in a cancelled state
	select {
	case <-ctx.Done():
		t.Fatalf("Context is cancelled/finished when it shouldn't be")
	default:
		// Do nothing, this means the context is correctly alive
	}
}
