package aws

import (
	"log"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsEMRInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRInstanceGroupCreate,
		Read:   resourceAwsEMRInstanceGroupRead,
		Update: resourceAwsEMRInstanceGroupUpdate,
		Delete: resourceAwsEMRInstanceGroupDelete,
		Schema: map[string]*schema.Schema{
			"cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_count": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"running_instance_count": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsEMRInstanceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	clusterId := d.Get("cluster_id").(string)
	instanceType := d.Get("instance_type").(string)
	instanceCount := d.Get("instance_count").(int)
	groupName := d.Get("name").(string)

	params := &emr.AddInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupConfig{
			{
				InstanceRole:  aws.String("TASK"),
				InstanceCount: aws.Int64(int64(instanceCount)),
				InstanceType:  aws.String(instanceType),
				Name:          aws.String(groupName),
			},
		},
		JobFlowId: aws.String(clusterId),
	}

	log.Printf("[DEBUG] Creating EMR task group params: %s", params)
	resp, err := conn.AddInstanceGroups(params)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Created EMR task group finished: %#v", resp)
	d.SetId(*resp.InstanceGroupIds[0])

	return nil
}

func resourceAwsEMRInstanceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn
	req := &emr.ListInstanceGroupsInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	}

	var group *emr.InstanceGroup
	marker := aws.String("intitial")
	for marker != nil {
		log.Printf("[DEBUG] EMR Cluster Instance Marker: %s", *marker)
		respGrps, errGrps := conn.ListInstanceGroups(req)
		if errGrps != nil {
			return fmt.Errorf("[ERR] Error reading EMR cluster (%s): %s", d.Id(), errGrps)
		}
		if respGrps == nil {
			return fmt.Errorf("[ERR] Error reading EMR Instance Group (%s)", d.Id())
		}

		instanceGroups := respGrps.InstanceGroups
		log.Printf("\n@@@\n[DEBUG] Instance Groups: \n%s\n@@@\n", instanceGroups)

		for _, ig := range respGrps.InstanceGroups {
			if d.Id() == *ig.Id {
				group = ig
				break
			}
		}
		if group != nil {
			// break out of marker loop
			break
		}
		marker = respGrps.Marker
	}

	if group == nil {
		log.Printf("[DEBUG] EMR Instance Group (%s) not found, removing", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", group.Name)
	d.Set("instance_count", group.RequestedInstanceCount)
	d.Set("running_instance_count", group.RunningInstanceCount)
	d.Set("instance_type", group.InstanceType)
	if group.Status != nil && group.Status.State != nil {
		d.Set("status", group.Status.State)
	}

	return nil
}

func resourceAwsEMRInstanceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	log.Printf("[DEBUG] Modify EMR task group")
	instanceCount := d.Get("instance_count").(int)

	if d.HasChange("name") {
		return fmt.Errorf("[WARN] Error updating task group, change name is not supported by api")
	}

	params := &emr.ModifyInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupModifyConfig{
			{
				InstanceGroupId: aws.String(d.Id()),
				InstanceCount:   aws.Int64(int64(instanceCount)),
			},
		},
	}
	resp, err := conn.ModifyInstanceGroups(params)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Modify EMR task group finished: %s", resp)

	return nil
}

func resourceAwsEMRInstanceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] AWS EMR Instance Group does not support DELETE; resizing cluster to zero before removing from state")
	conn := meta.(*AWSClient).emrconn
	params := &emr.ModifyInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupModifyConfig{
			{
				InstanceGroupId: aws.String(d.Id()),
				InstanceCount:   aws.Int64(0),
			},
		},
	}

	_, err := conn.ModifyInstanceGroups(params)
	if err != nil {
		return err
	}
	return nil
}
