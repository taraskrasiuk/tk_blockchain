package main

import (
	"errors"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		pbKey, ok := ctx.GetConfig("luko_tutor:publicKey")
		if !ok {
			return errors.New("the public key has not being defined")
		}
		name, _ := ctx.GetConfig("luko_tutor:name")
		if !ok {
			return errors.New("the application name has not being defined")
		}

		keyPair, err := ec2.NewKeyPair(ctx, name, &ec2.KeyPairArgs{
			PublicKey: pulumi.String(pbKey),
		})
		if err != nil {
			return err
		}
		// create security group
		gr, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(80),
					ToPort:     pulumi.Int(80),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(8080),
					ToPort:     pulumi.Int(8080),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
		})
		if err != nil {
			return err
		}
		// ami
		mostRecent := true
		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			Filters: []ec2.GetAmiFilter{
				{
					Name:   "name",
					Values: []string{"al2023-ami-*-x86_64"},
				},
			},
			Owners:     []string{"137112412989"},
			MostRecent: &mostRecent,
		})
		if err != nil {
			return err
		}
		// Create a simple web server using the startup script for the instance.
		srv, err := ec2.NewInstance(ctx, name, &ec2.InstanceArgs{
			Tags:                pulumi.StringMap{"Name": pulumi.String(name)},
			InstanceType:        pulumi.String("t2.micro"), // t2.micro is available in the AWS free tier.
			VpcSecurityGroupIds: pulumi.StringArray{gr.ID()},
			Ami:                 pulumi.String(ami.Id),
			KeyName:             keyPair.KeyName,
		})

		// Export the resulting server's IP address and DNS name.
		ctx.Export("publicIp", srv.PublicIp)
		ctx.Export("publicHostName", srv.PublicDns)

		return nil
	})
}
